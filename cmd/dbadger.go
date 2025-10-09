package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	badger "github.com/dgraph-io/badger/v4"
	badgeroptions "github.com/dgraph-io/badger/v4/options"
)

var (
	bgrdb         *badger.DB
	cacheCounters map[string]uint64 = make(map[string]uint64)
)

func badgerConnect() *badger.DB {
	datadir := ToUnixSlash(filepath.Join(DataDir, "fbin"))
	MakeDirs(datadir)
	DebugInfo("badgerConnect", datadir)
	numCompactors := runtime.NumCPU()
	if numCompactors < 4 {
		numCompactors = 4
	}
	if numCompactors > 16 {
		numCompactors = 16
	}
	opts := badger.DefaultOptions(datadir)
	opts.Dir = datadir
	opts.ValueDir = datadir
	opts.BaseTableSize = 256 << 20
	opts.NumVersionsToKeep = 1
	opts.NumCompactors = numCompactors
	opts.Compression = badgeroptions.ZSTD
	opts.ValueLogFileSize = 2<<30 - 64<<20
	opts.SyncWrites = false
	opts.ValueThreshold = 32
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
			//DebugInfo("badgerSetKV", "SKIP as exists")
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
			//DebugInfo("badgerSetV", "SKIP as exists")
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

func badgerCount(prefix string) uint64 {
	// saving memory
	if len(cacheCounters) > 100 {
		for k, _ := range cacheCounters {
			delete(cacheCounters, k)
		}
	}

	cacheKey := strings.Join([]string{"keycount", prefix}, "_")
	cacheVersion, ok := cacheCounters["cacheVersion"]
	if ok {
		if cacheVersion == bgrdb.MaxVersion() {
			val, ok := cacheCounters[cacheKey]
			if ok {
				DebugInfo("badgerCount", "cache HIT: ", cacheKey)
				return val
			}
		} else {
			for k, _ := range cacheCounters {
				delete(cacheCounters, k)
			}
		}
	}

	counter := uint64(0)
	bgrdb.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 1000
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		prefixByte := []byte(prefix)
		for it.Seek(prefixByte); it.ValidForPrefix(prefixByte); it.Next() {
			counter++
		}
		return nil
	})
	cacheCounters["cacheVersion"] = bgrdb.MaxVersion()
	cacheCounters[cacheKey] = counter

	return counter
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

func badgerSync() error {
	err := bgrdb.Sync()
	PrintError("badgerSync", err)
	return err
}

func badgerBackup(fpath string, fsince uint64) error {
	doneFile := strings.Join([]string{fpath, "backup", "done"}, ".")
	RemoveFile(doneFile)
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func(fpath string, fsince uint64, bgrdb *badger.DB, doneFile string) {
		defer wg.Done()
		fpathTemp := strings.Join([]string{fpath, "ing"}, ".")
		ft, err := os.Create(fpathTemp)
		if err != nil {
			PrintError("Backup", err)
			return
		}
		defer ft.Close()

		lastVersion, err := bgrdb.Backup(ft, fsince)
		if err != nil {
			PrintError("Backup", err)
			return
		}
		ft.Close()

		lastVersionFile := ToUnixSlash(filepath.Join(filepath.Dir(fpath), "ver"))
		DebugInfo("Backup", lastVersion)
		err = WriteFile(lastVersionFile, []byte(Uint64ToString(lastVersion)))
		if err != nil {
			DebugWarn("Backup", err)
			return
		}
		ftarget := strings.Join([]string{fpath, fmt.Sprintf("_[%v_%v]", fsince, lastVersion), ".zstdb.bak"}, "")
		err = os.Rename(fpathTemp, ftarget)

		if err != nil {
			PrintError("Backup", err)
			return
		}

		WriteFile(doneFile, []byte(ftarget))

	}(fpath, fsince, bgrdb, doneFile)

	wg.Wait()

	df, err := os.Stat(doneFile)
	if err != nil {
		return NewError("backup failed")
	}

	if time.Since(df.ModTime()).Seconds() > 3 {
		return NewError("backup failed")
	}

	DebugInfo("badgerBackup", "complete")
	return nil
}

func badgerRestore(fpath string) error {
	DebugInfo("badgerRestore", "from: ", fpath)
	errorFile := strings.Join([]string{fpath, "restore", "error"}, ".")
	RemoveFile(errorFile)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func(fpath string, bgrdb *badger.DB, errorFile string) {
		defer wg.Done()
		ft, err := os.Open(fpath)
		if err != nil {
			PrintError("Restore", err)
			WriteFile(errorFile, []byte(err.Error()))
			return
		}
		defer ft.Close()

		err = bgrdb.Load(ft, 16)
		if err != nil {
			PrintError("Restore", err)
			WriteFile(errorFile, []byte(err.Error()))
			return
		}

	}(fpath, bgrdb, errorFile)

	wg.Wait()

	_, err := os.Stat(errorFile)
	if err != nil {
		DebugInfo("badgerRestore", "complete")
		return nil
	}

	errContent := ReadFile(errorFile)
	if errContent != nil {
		return NewError(string(errContent))
	}

	return nil
}

func BadgerRunValueLogGC() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
	again:
		DebugInfo("RunValueLogGC", 0.5)
		err := bgrdb.RunValueLogGC(0.5)
		if err == nil {
			time.Sleep(3 * time.Second)
			goto again
		} else {
			PrintError("BadgerRunValueLogGC", err)
		}
	}
}
