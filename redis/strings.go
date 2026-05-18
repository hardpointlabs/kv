package redis

import (
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/tidwall/redcon"
)

func setKey(conn redcon.Conn, db *badger.DB, key, value []byte) {
	err := db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry(rawKeyPrefix(key, currentDb(conn)), value).WithMeta(byte(RedisString))
		return txn.SetEntry(e)
	})
	if err != nil {
		conn.WriteError("ERR " + err.Error())
		return
	}
	conn.WriteString("OK")
}

func setKeyWithTTL(conn redcon.Conn, db *badger.DB, key, value []byte, ttlSec int) {
	expTime := time.Duration(ttlSec) * time.Second
	err := db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry(rawKeyPrefix(key, currentDb(conn)), value).WithTTL(expTime).WithMeta(byte(RedisString))
		return txn.SetEntry(e)
	})
	if err != nil {
		conn.WriteError("ERR " + err.Error())
		return
	}
	conn.WriteString("OK")
}

func getKey(conn redcon.Conn, db *badger.DB, key []byte) {
	_ = db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(rawKeyPrefix(key, currentDb(conn)))
		if err != nil {
			conn.WriteNull()
			return nil
		}
		if item.UserMeta() != byte(RedisString) {
			conn.WriteError("WRONGTYPE Operation against a key holding the wrong kind of value")
			return nil
		}
		var valCopy []byte
		err = item.Value(func(val []byte) error {
			valCopy = append([]byte{}, val...)
			return nil
		})
		if err != nil {
			conn.WriteError("ERR " + err.Error())
			return err
		}
		conn.WriteBulk(valCopy)
		return nil
	})
}

func getSet(conn redcon.Conn, db *badger.DB, key, value []byte) {
	_ = db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get(rawKeyPrefix(key, currentDb(conn)))
		if err != nil {
			e := badger.NewEntry(rawKeyPrefix(key, currentDb(conn)), value).WithMeta(byte(RedisString))
			err = txn.SetEntry(e)
			if err != nil {
				conn.WriteError("ERR " + err.Error())
				return err
			}
			conn.WriteNull()
			return nil
		}
		var valCopy []byte
		err = item.Value(func(val []byte) error {
			valCopy = append([]byte{}, val...)
			return nil
		})
		if err != nil {
			conn.WriteError("ERR " + err.Error())
			return err
		}

		e := badger.NewEntry(rawKeyPrefix(key, currentDb(conn)), value).WithMeta(byte(RedisString))
		err = txn.SetEntry(e)
		if err != nil {
			conn.WriteError("ERR " + err.Error())
			return err
		}
		conn.WriteBulk(valCopy)
		return nil
	})
}

func getDel(conn redcon.Conn, db *badger.DB, key []byte) {
	_ = db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get(rawKeyPrefix(key, currentDb(conn)))
		if err != nil {
			conn.WriteNull()
			return nil
		}
		if item.UserMeta() != byte(RedisString) {
			conn.WriteError("WRONGTYPE Operation against a key holding the wrong kind of value")
			return nil
		}
		var valCopy []byte
		err = item.Value(func(val []byte) error {
			valCopy = append([]byte{}, val...)
			return nil
		})
		if err != nil {
			conn.WriteError("ERR " + err.Error())
			return err
		}

		err = txn.Delete(rawKeyPrefix(key, currentDb(conn)))
		if err != nil {
			conn.WriteNull()
			return nil
		}
		conn.WriteBulk(valCopy)
		return nil
	})
}

func strlenKey(conn redcon.Conn, db *badger.DB, key []byte) {
	_ = db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(rawKeyPrefix(key, currentDb(conn)))
		if err != nil {
			conn.WriteInt(0)
			return nil
		}
		if item.UserMeta() != byte(RedisString) {
			conn.WriteError("WRONGTYPE Operation against a key holding the wrong kind of value")
			return nil
		}
		var valCopy []byte
		err = item.Value(func(val []byte) error {
			valCopy = append([]byte{}, val...)
			return nil
		})
		if err != nil {
			conn.WriteError("ERR " + err.Error())
			return err
		}
		conn.WriteInt(len(valCopy))
		return nil
	})
}

func substrKey(conn redcon.Conn, db *badger.DB, key []byte, start, end int) {
	_ = db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(rawKeyPrefix(key, currentDb(conn)))
		if err != nil {
			conn.WriteNull()
			return nil
		}
		if item.UserMeta() != byte(RedisString) {
			conn.WriteError("WRONGTYPE Operation against a key holding the wrong kind of value")
			return nil
		}
		var valCopy []byte
		err = item.Value(func(val []byte) error {
			valCopy = append([]byte{}, val...)
			return nil
		})
		if err != nil {
			conn.WriteError("ERR " + err.Error())
			return nil
		}
		if start < 0 {
			start = len(valCopy) + start
		}
		if end < 0 {
			end = len(valCopy) + end
		}
		if start < 0 {
			start = 0
		}
		if end >= len(valCopy) {
			end = len(valCopy) - 1
		}
		if start > end || start >= len(valCopy) {
			conn.WriteBulk([]byte{})
			return nil
		}
		conn.WriteBulk(valCopy[start : end+1])
		return nil
	})
}

func setNX(conn redcon.Conn, db *badger.DB, key, value []byte) {
	_ = db.Update(func(txn *badger.Txn) error {
		var set = false
		_, err := txn.Get(rawKeyPrefix(key, currentDb(conn)))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				e := badger.NewEntry(rawKeyPrefix(key, currentDb(conn)), value).WithMeta(byte(RedisString))
				err = txn.SetEntry(e)
				if err != nil {
					return err
				}
				set = true
			} else {
				return err
			}
		}

		if err != nil {
			conn.WriteError("ERR " + err.Error())
			return nil
		}
		if set {
			conn.WriteInt(1)
		} else {
			conn.WriteInt(0)
		}
		return nil
	})
}
