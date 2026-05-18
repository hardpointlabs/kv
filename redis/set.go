package redis

import (
	"bytes"
	"encoding/binary"
	"math/rand/v2"
	"strconv"

	"github.com/dgraph-io/badger/v4"
	"github.com/tidwall/redcon"
)

// Internal key format:  -{db}:{setname}\x00{member}
// Sentinel key format:  {db}:{setname}           (value = 4-byte uint32 count)

func internalSetKey(setName, member []byte, dbSlot int) []byte {
	prefix := append([]byte(internalPrefix), []byte(strconv.Itoa(dbSlot)+prefixSeparator)...)
	prefix = append(prefix, setName...)
	prefix = append(prefix, 0)
	return append(prefix, member...)
}

func membersPrefix(setName []byte, dbSlot int) []byte {
	prefix := append([]byte(internalPrefix), []byte(strconv.Itoa(dbSlot)+prefixSeparator)...)
	prefix = append(prefix, setName...)
	return append(prefix, 0)
}

func memberFromInternalKey(key []byte) []byte {
	idx := bytes.LastIndexByte(key, 0)
	if idx < 0 {
		return nil
	}
	return key[idx+1:]
}

func writeSetCount(txn *badger.Txn, setName []byte, count uint32, dbSlot int) error {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, count)
	return txn.SetEntry(badger.NewEntry(rawKeyPrefixWithDb(setName, dbSlot), buf).WithMeta(byte(RedisSet)))
}

func readSetCount(txn *badger.Txn, setName []byte, dbSlot int) (uint32, error) {
	item, err := txn.Get(rawKeyPrefixWithDb(setName, dbSlot))
	if err != nil {
		return 0, err
	}
	var count uint32
	if err := item.Value(func(val []byte) error {
		count = binary.BigEndian.Uint32(val)
		return nil
	}); err != nil {
		return 0, err
	}
	return count, nil
}

func loadSetMembers(txn *badger.Txn, setName []byte, dbSlot int) (map[string]struct{}, error) {
	prefix := membersPrefix(setName, dbSlot)
	opts := badger.DefaultIteratorOptions
	opts.Prefix = prefix
	it := txn.NewIterator(opts)
	defer it.Close()

	members := make(map[string]struct{})
	for it.Rewind(); it.Valid(); it.Next() {
		item := it.Item()
		_ = item.Value(func(val []byte) error {
			members[string(val)] = struct{}{}
			return nil
		})
	}
	return members, nil
}

func clearSet(txn *badger.Txn, setName []byte, dbSlot int) error {
	prefix := membersPrefix(setName, dbSlot)
	opts := badger.DefaultIteratorOptions
	opts.Prefix = prefix
	it := txn.NewIterator(opts)
	defer it.Close()
	for it.Rewind(); it.Valid(); it.Next() {
		if err := txn.Delete(it.Item().KeyCopy(nil)); err != nil {
			return err
		}
	}
	return txn.Delete(rawKeyPrefixWithDb(setName, dbSlot))
}

func sadd(txn *badger.Txn, conn redcon.Conn, key []byte, members ...[]byte) (int, error) {
	dbSlot := 0
	if conn != nil {
		dbSlot = currentDb(conn)
	}

	var added int
	var count uint32

	item, err := txn.Get(rawKeyPrefixWithDb(key, dbSlot))
	if err == badger.ErrKeyNotFound {
		count = 0
	} else if err != nil {
		return 0, err
	} else {
		if err := item.Value(func(val []byte) error {
			count = binary.BigEndian.Uint32(val)
			return nil
		}); err != nil {
			return 0, err
		}
	}

	for _, member := range members {
		_, err := txn.Get(internalSetKey(key, member, dbSlot))
		if err == badger.ErrKeyNotFound {
			if err := txn.SetEntry(badger.NewEntry(internalSetKey(key, member, dbSlot), member).WithMeta(byte(RedisSet))); err != nil {
				return added, err
			}
			added++
			count++
		} else if err != nil {
			return added, err
		}
	}

	return added, writeSetCount(txn, key, count, dbSlot)
}

func srem(txn *badger.Txn, conn redcon.Conn, key []byte, members ...[]byte) (int, error) {
	dbSlot := 0
	if conn != nil {
		dbSlot = currentDb(conn)
	}

	count, err := readSetCount(txn, key, dbSlot)
	if err == badger.ErrKeyNotFound {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	var removed int
	for _, member := range members {
		internalKey := internalSetKey(key, member, dbSlot)
		_, err := txn.Get(internalKey)
		if err == badger.ErrKeyNotFound {
			continue
		}
		if err != nil {
			return removed, err
		}
		if err := txn.Delete(internalKey); err != nil {
			return removed, err
		}
		removed++
		count--
	}

	if count == 0 {
		return removed, txn.Delete(rawKeyPrefixWithDb(key, dbSlot))
	}
	return removed, writeSetCount(txn, key, count, dbSlot)
}

func scard(txn *badger.Txn, conn redcon.Conn, key []byte) (int, error) {
	dbSlot := 0
	if conn != nil {
		dbSlot = currentDb(conn)
	}

	count, err := readSetCount(txn, key, dbSlot)
	if err == badger.ErrKeyNotFound {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func smembers(txn *badger.Txn, conn redcon.Conn, key []byte) ([][]byte, error) {
	dbSlot := 0
	if conn != nil {
		dbSlot = currentDb(conn)
	}

	_, err := txn.Get(rawKeyPrefixWithDb(key, dbSlot))
	if err == badger.ErrKeyNotFound {
		return [][]byte{}, nil
	}
	if err != nil {
		return nil, err
	}

	prefix := membersPrefix(key, dbSlot)
	opts := badger.DefaultIteratorOptions
	opts.Prefix = prefix
	it := txn.NewIterator(opts)
	defer it.Close()

	var members [][]byte
	for it.Rewind(); it.Valid(); it.Next() {
		item := it.Item()
		_ = item.Value(func(val []byte) error {
			members = append(members, append([]byte{}, val...))
			return nil
		})
	}
	return members, nil
}

func sismember(txn *badger.Txn, conn redcon.Conn, key, member []byte) (bool, error) {
	dbSlot := 0
	if conn != nil {
		dbSlot = currentDb(conn)
	}

	_, err := txn.Get(internalSetKey(key, member, dbSlot))
	if err == badger.ErrKeyNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func spop(txn *badger.Txn, conn redcon.Conn, key []byte) ([]byte, error) {
	dbSlot := 0
	if conn != nil {
		dbSlot = currentDb(conn)
	}

	count, err := readSetCount(txn, key, dbSlot)
	if err == badger.ErrKeyNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	prefix := membersPrefix(key, dbSlot)
	opts := badger.DefaultIteratorOptions
	opts.Prefix = prefix
	it := txn.NewIterator(opts)
	defer it.Close()

	idx := rand.IntN(int(count))
	var member []byte
	var found int
	for it.Rewind(); it.Valid() && found <= idx; it.Next() {
		if found == idx {
			k := it.Item().KeyCopy(nil)
			member = memberFromInternalKey(k)
			if err := txn.Delete(k); err != nil {
				return nil, err
			}
		}
		found++
	}

	count--
	if count == 0 {
		return member, txn.Delete(rawKeyPrefixWithDb(key, dbSlot))
	}
	return member, writeSetCount(txn, key, count, dbSlot)
}

func srandmember(txn *badger.Txn, conn redcon.Conn, key []byte, count int) ([][]byte, error) {
	dbSlot := 0
	if conn != nil {
		dbSlot = currentDb(conn)
	}

	all, err := loadSetMembers(txn, key, dbSlot)
	if err == badger.ErrKeyNotFound {
		return [][]byte{}, nil
	}
	if err != nil {
		return nil, err
	}

	members := make([][]byte, 0, len(all))
	for m := range all {
		members = append(members, []byte(m))
	}

	if count == 0 || len(members) == 0 {
		return [][]byte{}, nil
	}

	if count > 0 && count >= len(members) {
		return members, nil
	}

	if count > 0 {
		perm := rand.Perm(len(members))
		result := make([][]byte, count)
		for i := 0; i < count; i++ {
			result[i] = members[perm[i]]
		}
		return result, nil
	}

	count = -count
	result := make([][]byte, count)
	for i := 0; i < count; i++ {
		result[i] = members[rand.IntN(len(members))]
	}
	return result, nil
}

func smove(txn *badger.Txn, conn redcon.Conn, src, dst, member []byte) (bool, error) {
	if bytes.Equal(src, dst) {
		_, err := txn.Get(internalSetKey(src, member, currentDb(conn)))
		if err == badger.ErrKeyNotFound {
			return false, nil
		}
		return err == nil, err
	}

	dbSlot := 0
	if conn != nil {
		dbSlot = currentDb(conn)
	}

	srcKey := internalSetKey(src, member, dbSlot)
	_, err := txn.Get(srcKey)
	if err == badger.ErrKeyNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	if err := txn.Delete(srcKey); err != nil {
		return false, err
	}

	srcCount, err := readSetCount(txn, src, dbSlot)
	if err != nil {
		return false, err
	}
	srcCount--
	if srcCount == 0 {
		if err := txn.Delete(rawKeyPrefixWithDb(src, dbSlot)); err != nil {
			return false, err
		}
	} else {
		if err := writeSetCount(txn, src, srcCount, dbSlot); err != nil {
			return false, err
		}
	}

	dstKey := internalSetKey(dst, member, dbSlot)
	_, err = txn.Get(dstKey)
	if err == badger.ErrKeyNotFound {
		if err := txn.SetEntry(badger.NewEntry(dstKey, member).WithMeta(byte(RedisSet))); err != nil {
			return false, err
		}
		dstCount, err := readSetCount(txn, dst, dbSlot)
		if err == badger.ErrKeyNotFound {
			return true, writeSetCount(txn, dst, 1, dbSlot)
		}
		if err != nil {
			return false, err
		}
		dstCount++
		return true, writeSetCount(txn, dst, dstCount, dbSlot)
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

func sdiff(txn *badger.Txn, conn redcon.Conn, keys ...[]byte) ([][]byte, error) {
	if len(keys) == 0 {
		return [][]byte{}, nil
	}
	dbSlot := 0
	if conn != nil {
		dbSlot = currentDb(conn)
	}

	result, err := loadSetMembers(txn, keys[0], dbSlot)
	if err != nil {
		return nil, err
	}

	for _, key := range keys[1:] {
		other, err := loadSetMembers(txn, key, dbSlot)
		if err != nil {
			return nil, err
		}
		for m := range other {
			delete(result, m)
		}
	}

	var members [][]byte
	for m := range result {
		members = append(members, []byte(m))
	}
	return members, nil
}

func sinter(txn *badger.Txn, conn redcon.Conn, keys ...[]byte) ([][]byte, error) {
	if len(keys) == 0 {
		return [][]byte{}, nil
	}
	dbSlot := 0
	if conn != nil {
		dbSlot = currentDb(conn)
	}

	result, err := loadSetMembers(txn, keys[0], dbSlot)
	if err != nil {
		return nil, err
	}

	for _, key := range keys[1:] {
		other, err := loadSetMembers(txn, key, dbSlot)
		if err != nil {
			return nil, err
		}
		for m := range result {
			if _, ok := other[m]; !ok {
				delete(result, m)
			}
		}
	}

	var members [][]byte
	for m := range result {
		members = append(members, []byte(m))
	}
	return members, nil
}

func sunion(txn *badger.Txn, conn redcon.Conn, keys ...[]byte) ([][]byte, error) {
	if len(keys) == 0 {
		return [][]byte{}, nil
	}
	dbSlot := 0
	if conn != nil {
		dbSlot = currentDb(conn)
	}

	result := make(map[string]struct{})
	for _, key := range keys {
		other, err := loadSetMembers(txn, key, dbSlot)
		if err != nil {
			return nil, err
		}
		for m := range other {
			result[m] = struct{}{}
		}
	}

	var members [][]byte
	for m := range result {
		members = append(members, []byte(m))
	}
	return members, nil
}

func storeSetResult(txn *badger.Txn, conn redcon.Conn, dest []byte, members [][]byte) (int, error) {
	dbSlot := 0
	if conn != nil {
		dbSlot = currentDb(conn)
	}

	_ = clearSet(txn, dest, dbSlot)

	for _, m := range members {
		if err := txn.SetEntry(badger.NewEntry(internalSetKey(dest, m, dbSlot), m).WithMeta(byte(RedisSet))); err != nil {
			return 0, err
		}
	}

	return len(members), writeSetCount(txn, dest, uint32(len(members)), dbSlot)
}

func sdiffstore(txn *badger.Txn, conn redcon.Conn, dest []byte, keys ...[]byte) (int, error) {
	result, err := sdiff(txn, conn, keys...)
	if err != nil {
		return 0, err
	}
	return storeSetResult(txn, conn, dest, result)
}

func sinterstore(txn *badger.Txn, conn redcon.Conn, dest []byte, keys ...[]byte) (int, error) {
	result, err := sinter(txn, conn, keys...)
	if err != nil {
		return 0, err
	}
	return storeSetResult(txn, conn, dest, result)
}

func sunionstore(txn *badger.Txn, conn redcon.Conn, dest []byte, keys ...[]byte) (int, error) {
	result, err := sunion(txn, conn, keys...)
	if err != nil {
		return 0, err
	}
	return storeSetResult(txn, conn, dest, result)
}
