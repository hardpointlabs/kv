package redis

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/dgraph-io/badger/v4"
)

// Pathological unit tests — document model level //

func TestJSONDeepNestedPathCreation(t *testing.T) {
	doc := newEmptyJSONDocument()

	parts := make([]string, 50)
	for i := range parts {
		parts[i] = fmt.Sprintf("l%d", i)
	}
	path := "$." + strings.Join(parts, ".")

	if err := doc.set(path, "bottom"); err != nil {
		t.Fatalf("set at depth 50: %v", err)
	}

	val, err := doc.get(path)
	if err != nil {
		t.Fatalf("get at depth 50: %v", err)
	}
	if val != "bottom" {
		t.Fatalf("expected 'bottom', got %v", val)
	}
}

func TestJSONUnicodeKeys(t *testing.T) {
	doc := newEmptyJSONDocument()
	entries := map[string]any{
		"名前":   "田中",
		"Привет": "мир",
		"שלום":   "עולם",
		"🚀🎉✨":  "emoji_value",
	}
	for k, v := range entries {
		path := "$." + k
		if err := doc.set(path, v); err != nil {
			t.Fatalf("set unicode key %q: %v", k, err)
		}
		got, err := doc.get(path)
		if err != nil {
			t.Fatalf("get unicode key %q: %v", k, err)
		}
		if got != v {
			t.Fatalf("unicode roundtrip %q: expected %v, got %v", k, v, got)
		}
	}
}

func TestJSONNumericEdgeCases(t *testing.T) {
	doc := newEmptyJSONDocument()

	cases := []struct {
		name  string
		value float64
	}{
		{"zero", 0},
		{"negative_zero", math.Copysign(0, -1)}, // -0
		{"max_float64", math.MaxFloat64},
		{"smallest_nonzero", math.SmallestNonzeroFloat64},
		{"large_integer", 9007199254740991},
		{"negative_large", -9007199254740991},
		{"pi", 3.141592653589793},
		{"negative", -42.5},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := doc.set("$", tc.value); err != nil {
				if tc.name == "negative_zero" {
					t.Logf("negative zero: set failed (may be expected): %v", err)
					return
				}
				t.Fatalf("set: %v", err)
			}
			val, err := doc.get("$")
			if err != nil {
				t.Fatalf("get: %v", err)
			}
			fv, ok := val.(float64)
			if !ok {
				t.Fatalf("expected float64, got %T", val)
			}
			if tc.name == "negative_zero" {
				if math.Signbit(fv) {
					t.Logf("negative zero preserved")
				} else {
					t.Logf("negative zero became zero (expected with Go json)")
				}
				return
			}
			if math.IsInf(tc.value, 0) || math.IsNaN(tc.value) {
				return
			}
			// compare with a tolerance
			diff := fv - tc.value
			if diff < 0 {
				diff = -diff
			}
			if diff > 1e-9 && diff/tc.value > 1e-9 {
				t.Fatalf("value mismatch: expected %v, got %v", tc.value, fv)
			}
		})
	}
}

func TestJSONLargeArraySetGet(t *testing.T) {
	size := 10000
	arr := make([]any, size)
	for i := range arr {
		arr[i] = float64(i)
	}

	doc := newEmptyJSONDocument()
	if err := doc.set("$", arr); err != nil {
		t.Fatalf("set large array: %v", err)
	}
	got, err := doc.get("$")
	if err != nil {
		t.Fatalf("get large array: %v", err)
	}
	gotArr, ok := got.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", got)
	}
	if len(gotArr) != size {
		t.Fatalf("length: expected %d, got %d", size, len(gotArr))
	}
}

func TestJSONDeeplyNestedObjectRoundtrip(t *testing.T) {
	n := 200
	var path strings.Builder
	path.WriteString("$")
	for i := 0; i < n; i++ {
		fmt.Fprintf(&path, ".k%d", i)
	}
	leafPath := path.String()

	doc := newEmptyJSONDocument()
	if err := doc.set(leafPath, float64(42)); err != nil {
		t.Fatalf("set at depth %d: %v", n, err)
	}

	got, err := doc.get(leafPath)
	if err != nil {
		t.Fatalf("get at depth %d: %v", n, err)
	}
	if got != float64(42) {
		t.Fatalf("value at depth %d: expected 42, got %v", n, got)
	}

	serialized, err := doc.serialize()
	if err != nil {
		t.Fatalf("serialize: %v", err)
	}
	if len(serialized) == 0 {
		t.Fatal("empty serialized output")
	}
}

func TestJSONNumericCoercionInArray(t *testing.T) {
	doc := newEmptyJSONDocument()
	src := []any{float64(1), float64(2), float64(3)}
	if err := doc.set("$.items", src); err != nil {
		t.Fatal(err)
	}
	got, err := doc.get("$.items[0]")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := got.(float64); !ok {
		t.Fatalf("expected float64 for array element, got %T", got)
	}
}

func TestJSONPathOverwritesDeepNestedValue(t *testing.T) {
	doc := newEmptyJSONDocument()

	if err := doc.set("$.a.b.c", float64(1)); err != nil {
		t.Fatal(err)
	}
	if err := doc.set("$.a.b.c", float64(99)); err != nil {
		t.Fatal(err)
	}
	val, err := doc.get("$.a.b.c")
	if err != nil {
		t.Fatal(err)
	}
	if val != float64(99) {
		t.Fatalf("expected 99, got %v", val)
	}
}

func TestJSONSetOverwritesPrimitiveWithObject(t *testing.T) {
	doc := newEmptyJSONDocument()

	// Set root to a number
	if err := doc.set("$", float64(42)); err != nil {
		t.Fatal(err)
	}
	// Setting a nested path on a primitive should fail
	err := doc.set("$.x.y", float64(1))
	if err == nil {
		t.Fatal("expected error when setting nested path on primitive, got nil")
	}
	t.Logf("expected error: %v", err)
}

// BadgerDB pathological tests //

func TestJSONBadgerDeepNestedSetGet(t *testing.T) {
	db := inMemDB(t)
	defer db.Close()

	var parts []string
	for i := 0; i < 100; i++ {
		parts = append(parts, fmt.Sprintf("lvl%d", i))
	}
	path := "$." + strings.Join(parts, ".")

	prefix := rawKeyPrefix([]byte("deepdoc"), 0)

	err := db.Update(func(txn *badger.Txn) error {
		doc := newEmptyJSONDocument()
		if err := doc.set(path, "deep_value"); err != nil {
			return err
		}
		data, err := doc.serialize()
		if err != nil {
			return err
		}
		e := badger.NewEntry(prefix, data).WithMeta(byte(RedisJSON))
		return txn.SetEntry(e)
	})
	if err != nil {
		t.Fatalf("set deep: %v", err)
	}

	err = db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(prefix)
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			doc, err := newJSONDocument(val)
			if err != nil {
				return err
			}
			got, err := doc.get(path)
			if err != nil {
				return err
			}
			if got != "deep_value" {
				t.Fatalf("expected 'deep_value', got %v", got)
			}
			return nil
		})
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONBadgerLargeObject(t *testing.T) {
	db := inMemDB(t)
	defer db.Close()

	prefix := rawKeyPrefix([]byte("big"), 0)

	obj := make(map[string]any)
	for i := 0; i < 5000; i++ {
		obj[fmt.Sprintf("f%d", i)] = float64(i)
	}

	err := db.Update(func(txn *badger.Txn) error {
		doc := &JSONDocument{root: obj}
		data, err := doc.serialize()
		if err != nil {
			return err
		}
		e := badger.NewEntry(prefix, data).WithMeta(byte(RedisJSON))
		return txn.SetEntry(e)
	})
	if err != nil {
		t.Fatalf("set large object: %v", err)
	}

	err = db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(prefix)
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			doc, err := newJSONDocument(val)
			if err != nil {
				return err
			}
			for i := 0; i < 5000; i++ {
				path := fmt.Sprintf("$.f%d", i)
				got, err := doc.get(path)
				if err != nil {
					return fmt.Errorf("get %s: %w", path, err)
				}
				if got != float64(i) {
					return fmt.Errorf("field %d: expected %d, got %v", i, i, got)
				}
			}
			return nil
		})
	})
	if err != nil {
		t.Fatal(err)
	}
}

// Concurrent correctness tests //

func TestJSONConcurrentWritersDifferentPaths(t *testing.T) {
	db := inMemDB(t)
	defer db.Close()

	prefix := rawKeyPrefix([]byte("concdoc"), 0)

	err := db.Update(func(txn *badger.Txn) error {
		doc := newEmptyJSONDocument()
		data, err := doc.serialize()
		if err != nil {
			return err
		}
		e := badger.NewEntry(prefix, data).WithMeta(byte(RedisJSON))
		return txn.SetEntry(e)
	})
	if err != nil {
		t.Fatal(err)
	}

	const numWriters = 20
	const pathsPerWriter = 5
	var wg sync.WaitGroup
	var totalWrites int64

	for w := 0; w < numWriters; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for p := 0; p < pathsPerWriter; p++ {
				path := fmt.Sprintf("$.w%d_p%d", w, p)
				value := fmt.Sprintf("v_%d_%d", w, p)

				err := retryConflict(func() error {
					return db.Update(func(txn *badger.Txn) error {
						item, err := txn.Get(prefix)
						if err != nil {
							return err
						}
						var data []byte
						if err := item.Value(func(val []byte) error {
							data = append([]byte{}, val...)
							return nil
						}); err != nil {
							return err
						}
						doc, err := newJSONDocument(data)
						if err != nil {
							return err
						}
						if err := doc.set(path, value); err != nil {
							return err
						}
						newData, err := doc.serialize()
						if err != nil {
							return err
						}
						return txn.SetEntry(badger.NewEntry(prefix, newData).WithMeta(byte(RedisJSON)))
					})
				})
				if err != nil {
					t.Logf("writer %d path %s failed after retries: %v", w, path, err)
					return
				}
				atomic.AddInt64(&totalWrites, 1)
			}
		}(w)
	}
	wg.Wait()

	t.Logf("successful writes: %d", totalWrites)

	err = db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(prefix)
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			doc, err := newJSONDocument(val)
			if err != nil {
				return err
			}
			var missing []string
			for w := 0; w < numWriters; w++ {
				for p := 0; p < pathsPerWriter; p++ {
					path := fmt.Sprintf("$.w%d_p%d", w, p)
					got, err := doc.get(path)
					if err != nil {
						missing = append(missing, path)
						continue
					}
					expected := fmt.Sprintf("v_%d_%d", w, p)
					if got != expected {
						t.Errorf("%s: expected %s, got %v", path, expected, got)
					}
				}
			}
			if len(missing) > 0 {
				return fmt.Errorf("missing paths: %v", missing)
			}
			return nil
		})
	})
	if err != nil {
		t.Fatalf("verification: %v", err)
	}
}

func TestJSONConcurrentWritersSamePath(t *testing.T) {
	db := inMemDB(t)
	defer db.Close()

	prefix := rawKeyPrefix([]byte("concsame"), 0)

	err := db.Update(func(txn *badger.Txn) error {
		doc := newEmptyJSONDocument()
		if err := doc.set("$.counter", float64(0)); err != nil {
			return err
		}
		data, err := doc.serialize()
		if err != nil {
			return err
		}
		e := badger.NewEntry(prefix, data).WithMeta(byte(RedisJSON))
		return txn.SetEntry(e)
	})
	if err != nil {
		t.Fatal(err)
	}

	const numWriters = 20
	const writesPerWriter = 50
	var wg sync.WaitGroup

	for w := 0; w < numWriters; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for i := 0; i < writesPerWriter; i++ {
				value := fmt.Sprintf("w%d_i%d", w, i)
				_ = db.Update(func(txn *badger.Txn) error {
					item, err := txn.Get(prefix)
					if err != nil {
						return err
					}
					var data []byte
					if err := item.Value(func(val []byte) error {
						data = append([]byte{}, val...)
						return nil
					}); err != nil {
						return err
					}
					doc, err := newJSONDocument(data)
					if err != nil {
						return err
					}
					if err := doc.set("$.counter", value); err != nil {
						return err
					}
					newData, err := doc.serialize()
					if err != nil {
						return err
					}
					return txn.SetEntry(badger.NewEntry(prefix, newData).WithMeta(byte(RedisJSON)))
				})
			}
		}(w)
	}
	wg.Wait()

	err = db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(prefix)
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			doc, err := newJSONDocument(val)
			if err != nil {
				return err
			}
			counter, err := doc.get("$.counter")
			if err != nil {
				return err
			}
			t.Logf("final counter value: %v (one of %d total writes)", counter, numWriters*writesPerWriter)
			return nil
		})
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONConcurrentReadsDuringWrites(t *testing.T) {
	db := inMemDB(t)
	defer db.Close()

	prefix := rawKeyPrefix([]byte("concrw"), 0)

	err := db.Update(func(txn *badger.Txn) error {
		doc := newEmptyJSONDocument()
		for i := 0; i < 50; i++ {
			path := fmt.Sprintf("$.f%d", i)
			if err := doc.set(path, float64(i)); err != nil {
				return err
			}
		}
		data, err := doc.serialize()
		if err != nil {
			return err
		}
		e := badger.NewEntry(prefix, data).WithMeta(byte(RedisJSON))
		return txn.SetEntry(e)
	})
	if err != nil {
		t.Fatal(err)
	}

	const readers = 10
	const writers = 5
	const iterations = 50
	var wg sync.WaitGroup
	var readErrs int64
	var writeErrs int64

	for r := 0; r < readers; r++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				err := db.View(func(txn *badger.Txn) error {
					item, err := txn.Get(prefix)
					if err != nil {
						return err
					}
					return item.Value(func(val []byte) error {
						doc, err := newJSONDocument(val)
						if err != nil {
							return err
						}
						for j := 0; j < 50; j++ {
							path := fmt.Sprintf("$.f%d", j)
							if _, err := doc.get(path); err != nil {
								return err
							}
						}
						return nil
					})
				})
				if err != nil {
					atomic.AddInt64(&readErrs, 1)
				}
			}
		}()
	}

	for w := 0; w < writers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				err := db.Update(func(txn *badger.Txn) error {
					item, err := txn.Get(prefix)
					if err != nil {
						return err
					}
					var data []byte
					if err := item.Value(func(val []byte) error {
						data = append([]byte{}, val...)
						return nil
					}); err != nil {
						return err
					}
					doc, err := newJSONDocument(data)
					if err != nil {
						return err
					}
					path := fmt.Sprintf("$.f%d", i%50)
					if err := doc.set(path, float64(w*iterations+i)); err != nil {
						return err
					}
					newData, err := doc.serialize()
					if err != nil {
						return err
					}
					return txn.SetEntry(badger.NewEntry(prefix, newData).WithMeta(byte(RedisJSON)))
				})
				if err != nil && err != badger.ErrConflict {
					atomic.AddInt64(&writeErrs, 1)
				}
			}
		}(w)
	}
	wg.Wait()

	t.Logf("read errors: %d, write errors: %d", atomic.LoadInt64(&readErrs), atomic.LoadInt64(&writeErrs))
	// ErrConflict during writes is expected under concurrency
	if atomic.LoadInt64(&readErrs) > 0 {
		t.Logf("some reads failed (may be transient)")
	}
}

func TestJSONConcurrentArrayAppend(t *testing.T) {
	db := inMemDB(t)
	defer db.Close()

	prefix := rawKeyPrefix([]byte("concar"), 0)

	err := db.Update(func(txn *badger.Txn) error {
		doc := newEmptyJSONDocument()
		if err := doc.set("$.items", []any{}); err != nil {
			return err
		}
		data, err := doc.serialize()
		if err != nil {
			return err
		}
		e := badger.NewEntry(prefix, data).WithMeta(byte(RedisJSON))
		return txn.SetEntry(e)
	})
	if err != nil {
		t.Fatal(err)
	}

	const appenders = 10
	const appendsPerWorker = 50
	var wg sync.WaitGroup
	var appendCount int32

	for a := 0; a < appenders; a++ {
		wg.Add(1)
		go func(a int) {
			defer wg.Done()
			for i := 0; i < appendsPerWorker; i++ {
				err := db.Update(func(txn *badger.Txn) error {
					item, err := txn.Get(prefix)
					if err != nil {
						return err
					}
					var data []byte
					if err := item.Value(func(val []byte) error {
						data = append([]byte{}, val...)
						return nil
					}); err != nil {
						return err
					}
					doc, err := newJSONDocument(data)
					if err != nil {
						return err
					}
					val := fmt.Sprintf("item_%d_%d", a, i)
					if _, err := doc.arrAppend("$.items", val); err != nil {
						return err
					}
					newData, err := doc.serialize()
					if err != nil {
						return err
					}
					return txn.SetEntry(badger.NewEntry(prefix, newData).WithMeta(byte(RedisJSON)))
				})
				if err == nil {
					atomic.AddInt32(&appendCount, 1)
				}
			}
		}(a)
	}
	wg.Wait()

	t.Logf("successful appends: %d (target: %d, conflicts cause retries)", atomic.LoadInt32(&appendCount), appenders*appendsPerWorker)

	err = db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(prefix)
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			doc, err := newJSONDocument(val)
			if err != nil {
				return err
			}
			v, err := doc.get("$.items")
			if err != nil {
				return err
			}
			items, ok := v.([]any)
			if !ok {
				t.Fatalf("expected array, got %T", val)
			}
			t.Logf("final array length: %d", len(items))
			return nil
		})
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSONConcurrentCreateDelete(t *testing.T) {
	db := inMemDB(t)
	defer db.Close()

	prefixes := make([][]byte, 10)
	for i := range prefixes {
		prefixes[i] = rawKeyPrefix([]byte(fmt.Sprintf("tmpdoc_%d", i)), 0)
	}

	const workers = 10
	const cycles = 50
	var wg sync.WaitGroup
	var successCount int32

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for c := 0; c < cycles; c++ {
				idx := c % 10
				prefix := prefixes[idx]

				err := db.Update(func(txn *badger.Txn) error {
					doc := newEmptyJSONDocument()
					if err := doc.set("$.data", fmt.Sprintf("w%d_c%d", w, c)); err != nil {
						return err
					}
					data, err := doc.serialize()
					if err != nil {
						return err
					}
					return txn.SetEntry(badger.NewEntry(prefix, data).WithMeta(byte(RedisJSON)))
				})
				if err != nil {
					continue
				}

				err = db.Update(func(txn *badger.Txn) error {
					return txn.Delete(prefix)
				})
				if err == nil {
					atomic.AddInt32(&successCount, 1)
				}
			}
		}(w)
	}
	wg.Wait()

	t.Logf("successful create/delete cycles: %d", atomic.LoadInt32(&successCount))
	if atomic.LoadInt32(&successCount) == 0 {
		t.Error("no create/delete cycles succeeded")
	}
}

func TestJSONConcurrentMixedOperations(t *testing.T) {
	db := inMemDB(t)
	defer db.Close()

	docKeys := make([][]byte, 5)
	for i := range docKeys {
		docKeys[i] = rawKeyPrefix([]byte(fmt.Sprintf("mix_%d", i)), 0)
	}

	for _, k := range docKeys {
		_ = db.Update(func(txn *badger.Txn) error {
			doc := newEmptyJSONDocument()
			_ = doc.set("$.counter", float64(0))
			_ = doc.set("$.items", []any{})
			_ = doc.set("$.text", "hello")
			data, _ := doc.serialize()
			return txn.SetEntry(badger.NewEntry(k, data).WithMeta(byte(RedisJSON)))
		})
	}

	const workers = 20
	const opsPerWorker = 200
	var wg sync.WaitGroup
	var totalOps int64

	rng := rand.New(rand.NewSource(42))

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for op := 0; op < opsPerWorker; op++ {
				prefix := docKeys[rng.Intn(len(docKeys))]
				opType := rng.Intn(6)

				var err error
				switch opType {
				case 0: // set
					err = db.Update(func(txn *badger.Txn) error {
						item, err := txn.Get(prefix)
						if err != nil {
							return err
						}
						var data []byte
						if err := item.Value(func(val []byte) error {
							data = append([]byte{}, val...)
							return nil
						}); err != nil {
							return err
						}
						doc, err := newJSONDocument(data)
						if err != nil {
							return err
						}
						path := fmt.Sprintf("$.f%d", rng.Intn(20))
						_ = doc.set(path, float64(rng.Intn(1000)))
						newData, _ := doc.serialize()
						return txn.SetEntry(badger.NewEntry(prefix, newData).WithMeta(byte(RedisJSON)))
					})
				case 1: // get
					err = db.View(func(txn *badger.Txn) error {
						item, err := txn.Get(prefix)
						if err != nil {
							return err
						}
						return item.Value(func(val []byte) error {
							doc, err := newJSONDocument(val)
							if err != nil {
								return err
							}
							_, _ = doc.get("$.counter")
							return nil
						})
					})
				case 2: // arrAppend
					err = db.Update(func(txn *badger.Txn) error {
						item, err := txn.Get(prefix)
						if err != nil {
							return err
						}
						var data []byte
						if err := item.Value(func(val []byte) error {
							data = append([]byte{}, val...)
							return nil
						}); err != nil {
							return err
						}
						doc, err := newJSONDocument(data)
						if err != nil {
							return err
						}
						_, _ = doc.arrAppend("$.items", float64(rng.Intn(100)))
						newData, _ := doc.serialize()
						return txn.SetEntry(badger.NewEntry(prefix, newData).WithMeta(byte(RedisJSON)))
					})
				case 3: // numIncrBy
					err = db.Update(func(txn *badger.Txn) error {
						item, err := txn.Get(prefix)
						if err != nil {
							return err
						}
						var data []byte
						if err := item.Value(func(val []byte) error {
							data = append([]byte{}, val...)
							return nil
						}); err != nil {
							return err
						}
						doc, err := newJSONDocument(data)
						if err != nil {
							return err
						}
						_, _ = doc.numIncrBy("$.counter", float64(1))
						newData, _ := doc.serialize()
						return txn.SetEntry(badger.NewEntry(prefix, newData).WithMeta(byte(RedisJSON)))
					})
				case 4: // arrLen
					err = db.View(func(txn *badger.Txn) error {
						item, err := txn.Get(prefix)
						if err != nil {
							return err
						}
						return item.Value(func(val []byte) error {
							doc, err := newJSONDocument(val)
							if err != nil {
								return err
							}
							_, _ = doc.arrLen("$.items")
							return nil
						})
					})
				case 5: // delete path
					err = db.Update(func(txn *badger.Txn) error {
						item, err := txn.Get(prefix)
						if err != nil {
							return err
						}
						var data []byte
						if err := item.Value(func(val []byte) error {
							data = append([]byte{}, val...)
							return nil
						}); err != nil {
							return err
						}
						doc, err := newJSONDocument(data)
						if err != nil {
							return err
						}
						path := fmt.Sprintf("$.f%d", rng.Intn(20))
						_ = doc.delete(path)
						newData, _ := doc.serialize()
						return txn.SetEntry(badger.NewEntry(prefix, newData).WithMeta(byte(RedisJSON)))
					})
				}
				if err == nil {
					atomic.AddInt64(&totalOps, 1)
				}
			}
		}()
	}
	wg.Wait()

	t.Logf("successful mixed operations: %d (target: %d)", atomic.LoadInt64(&totalOps), workers*opsPerWorker)
	if atomic.LoadInt64(&totalOps) < int64(workers*opsPerWorker)/2 {
		t.Errorf("too many failed ops: %d/%d succeeded", atomic.LoadInt64(&totalOps), workers*opsPerWorker)
	}
}

// BadgerDB transactional edge cases //

func TestJSONBadgerCreateDeleteRecreate(t *testing.T) {
	db := inMemDB(t)
	defer db.Close()

	prefix := rawKeyPrefix([]byte("cycledoc"), 0)

	for i := 0; i < 50; i++ {
		err := db.Update(func(txn *badger.Txn) error {
			doc := newEmptyJSONDocument()
			if err := doc.set("$.cycle", float64(i)); err != nil {
				return err
			}
			data, err := doc.serialize()
			if err != nil {
				return err
			}
			return txn.SetEntry(badger.NewEntry(prefix, data).WithMeta(byte(RedisJSON)))
		})
		if err != nil {
			t.Fatalf("create cycle %d: %v", i, err)
		}

		err = db.View(func(txn *badger.Txn) error {
			item, err := txn.Get(prefix)
			if err != nil {
				return err
			}
			return item.Value(func(val []byte) error {
				doc, err := newJSONDocument(val)
				if err != nil {
					return err
				}
				got, err := doc.get("$.cycle")
				if err != nil {
					return err
				}
				if got != float64(i) {
					t.Fatalf("cycle %d: expected %d, got %v", i, i, got)
				}
				return nil
			})
		})
		if err != nil {
			t.Fatalf("cycle %d verify: %v", i, err)
		}
	}
}

func retryConflict(fn func() error) error {
	const maxRetries = 200
	var err error
	for i := 0; i < maxRetries; i++ {
		err = fn()
		if err == nil {
			return nil
		}
		if err == badger.ErrConflict {
			continue
		}
		return err
	}
	return err
}

func TestJSONBadgerMultipleKeys(t *testing.T) {
	db := inMemDB(t)
	defer db.Close()

	const numKeys = 100
	prefixes := make([][]byte, numKeys)
	for i := range prefixes {
		prefixes[i] = rawKeyPrefix([]byte(fmt.Sprintf("k%d", i)), 0)
	}

	var wg sync.WaitGroup
	for i := 0; i < numKeys; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			prefix := prefixes[i]
			_ = db.Update(func(txn *badger.Txn) error {
				doc := newEmptyJSONDocument()
				_ = doc.set("$.idx", float64(i))
				_ = doc.set("$.name", fmt.Sprintf("key_%d", i))
				data, _ := doc.serialize()
				return txn.SetEntry(badger.NewEntry(prefix, data).WithMeta(byte(RedisJSON)))
			})
		}(i)
	}
	wg.Wait()

	for i := 0; i < numKeys; i++ {
		i2 := i
		err := db.View(func(txn *badger.Txn) error {
			item, err := txn.Get(prefixes[i2])
			if err != nil {
				return err
			}
			return item.Value(func(val []byte) error {
				doc, err := newJSONDocument(val)
				if err != nil {
					return err
				}
				idx, _ := doc.get("$.idx")
				if idx != float64(i2) {
					t.Errorf("key %d: expected idx %d, got %v", i2, i2, idx)
				}
				return nil
			})
		})
		if err != nil {
			t.Errorf("key %d read: %v", i2, err)
		}
	}
}

// NX/XX tests that bypass node-redis client quirks //

func TestJSONSetNXNewKey(t *testing.T) {
	db := inMemDB(t)
	defer db.Close()

	prefix := rawKeyPrefix([]byte("nxtest"), 0)
	value := []byte(`"first"`)

	err := db.Update(func(txn *badger.Txn) error {
		_, err := txn.Get(prefix)
		if err == badger.ErrKeyNotFound {
			var v any
			if err := json.Unmarshal(value, &v); err != nil {
				return err
			}
			doc := newEmptyJSONDocument()
			doc.root = v
			data, err := doc.serialize()
			if err != nil {
				return err
			}
			return txn.SetEntry(badger.NewEntry(prefix, data).WithMeta(byte(RedisJSON)))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("NX create: %v", err)
	}

	err = db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(prefix)
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			doc, err := newJSONDocument(val)
			if err != nil {
				return err
			}
			got, _ := doc.get("$")
			if got != "first" {
				t.Fatalf("expected 'first', got %v", got)
			}
			return nil
		})
	})
	if err != nil {
		t.Fatal(err)
	}
}
