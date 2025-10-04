package cmd

import (
	"path/filepath"
	"strings"

	badger "github.com/dgraph-io/badger/v4"
)

var (
	bgrdb *badger.DB
)

func badgerConnect() *badger.DB {
	datadir := ToUnixSlash(filepath.Join(DataDir, "fbin"))
	MakeDirs(datadir)
	DebugInfo("badgerConnect", datadir)
	opts := badger.DefaultOptions(datadir)
	opts.Dir = datadir
	opts.ValueDir = datadir
	opts.BaseTableSize = 256 << 20
	opts.NumVersionsToKeep = 1
	opts.SyncWrites = false
	opts.ValueThreshold = 16
	opts.CompactL0OnClose = true

	db, err := badger.Open(opts)
	FatalError("badgerConnect", err)
	return db
}

func badgerSave(key, val []byte) []byte {
	if int64(len(val)) > MaxUploadSize {
		DebugWarn("badgerSetKV", "val is oversized")
		return nil
	}

	if IsAllowUserKey {
		return badgerSetKV(key, val)
	}

	return badgerSetV(val)
}

func badgerSetKV(key, val []byte) []byte {
	if IsAnyNil(key, val) {
		DebugWarn("badgerSetKV", "key/val cannot be empty")
		return nil
	}

	err := bgrdb.Update(func(txn *badger.Txn) error {
		_, err := txn.Get(key)
		if err == nil && IsAllowOverWrite == false {
			DebugInfo("badgerSetKV", "SKIP as exists")
			return nil
		}
		err = txn.Set(key, ZstdBytes(val))
		PrintError("badgerSetKV", err)
		return err
	})
	if err != nil {
		return nil
	}
	return key
}

func badgerSetV(val []byte) (key []byte) {
	if val == nil {
		DebugWarn("badgerSetV", "val cannot be empty")
		return nil
	}

	key = SumBlake3(val)

	err := bgrdb.Update(func(txn *badger.Txn) error {
		_, err := txn.Get(key)
		if err == nil && IsAllowOverWrite == false {
			DebugInfo("badgerSetV", "SKIP as exists")
			return nil
		}
		err = txn.Set([]byte(key), ZstdBytes(val))
		PrintError("badgerSetV", err)
		return err
	})
	if err != nil {
		return nil
	}
	return key
}

func badgerGet(key []byte) (val []byte) {
	if key == nil {
		DebugWarn("badgerGet.10", "key cannot be empty")
		return nil
	}

	bgrdb.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			DebugWarn("badgerGet.20", err, ":", string(key))
			return nil
		}
		itemVal, err := item.ValueCopy(nil)
		PrintError("badgerGet.30", err)
		//DebugInfo("badgerGet", len(itemVal), " :", string(key))
		val, err = UnZstdBytes(itemVal)
		if err != nil {
			return err
		}
		return err
	})

	return val
}

func badgerDelete(key []byte) error {
	if key == nil {
		DebugWarn("badgerDelete", "key cannot be empty")
		return NewError("key cannot be empty")
	}

	err := bgrdb.Update(func(txn *badger.Txn) error {
		_, err := txn.Get(key)
		if err != nil {
			return nil
		}
		err = txn.Delete(key)
		PrintError("badgerDelete", err)
		return err
	})

	return err
}

func badgerList(prefix string, pageNum int) []string {
	var pageKeys []string
	if pageNum < 1 {
		pageNum = 1
	}

	pageSize := 1000
	bgrdb.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		skipRows := (pageNum - 1) * pageSize
		counter := 0
		prefixByte := []byte(prefix)
		for it.Seek(prefixByte); it.ValidForPrefix(prefixByte); it.Next() {
			if counter < skipRows {
				counter++
				continue
			}

			if len(pageKeys) >= pageSize {
				break
			}
			item := it.Item()
			k := string(item.Key())
			if strings.HasPrefix(k, prefix) {
				pageKeys = append(pageKeys, k)
			}
		}
		return nil
	})

	return pageKeys
}

func badgerExists(key []byte) bool {
	if key == nil {
		DebugWarn("badgerExists", "key cannot be empty")
		return false
	}

	err := bgrdb.View(func(txn *badger.Txn) error {
		_, err := txn.Get(key)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return false
	}

	return true
}
