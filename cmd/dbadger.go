package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

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

func badgerGet(key []byte) (val []byte, ver uint64) {
	if key == nil {
		DebugWarn("badgerGet.10", "key cannot be empty")
		return nil, 0
	}

	bgrdb.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			DebugWarn("badgerGet.20", err, ":", string(key))
			return nil
		}

		itemVal, err := item.ValueCopy(nil)
		if err != nil {
			PrintError("badgerGet.30", err)
			return err
		}

		ver = item.Version()

		val, err = UnZstdBytes(itemVal)
		if err != nil {
			return err
		}
		return err
	})

	return val, ver
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

func badgerExists(key []byte) uint64 {
	if key == nil {
		DebugWarn("badgerExists", "key cannot be empty")
		return 0
	}
	var verNum uint64
	err := bgrdb.View(func(txn *badger.Txn) error {
		it, err := txn.Get(key)
		if err != nil {
			return err
		}
		verNum = it.Version()
		return nil
	})
	if err != nil {
		return 0
	}

	return verNum
}

func badgerBackup(fpath string, fsince uint64) error {
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func(fpath string, fsince uint64, bgrdb *badger.DB) {
		defer wg.Done()
		ft, err := os.Create(fpath)
		if err != nil {
			PrintError("Backup", err)
			return
		}
		defer ft.Close()

		n, err := bgrdb.Backup(ft, fsince)
		if err != nil {
			PrintError("Backup", err)
			return
		}
		DebugInfo("Backup", n)
	}(fpath, fsince, bgrdb)

	wg.Wait()

	DebugInfo("badgerBackup", "complete")
	return nil
}

func badgerRestore(fpath string) error {
	DebugInfo("badgerRestore", "from: ", fpath)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func(fpath string, bgrdb *badger.DB) {
		defer wg.Done()
		ft, err := os.Open(fpath)
		if err != nil {
			PrintError("Restore", err)
			return
		}
		defer ft.Close()

		err = bgrdb.Load(ft, 16)
		if err != nil {
			PrintError("Restore", err)
			return
		}
	}(fpath, bgrdb)

	wg.Wait()

	DebugInfo("badgerRestore", "complete")
	return nil
}

func BadgerRunValueLogGC() {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
	again:
		DebugInfo("RunValueLogGC", 0.7)
		err := bgrdb.RunValueLogGC(0.7)
		if err == nil {
			goto again
		}
	}
}
