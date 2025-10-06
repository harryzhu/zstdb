package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/klauspost/compress/zstd"
	"github.com/zeebo/blake3"
)

func GetNowUnix() int64 {
	return time.Now().UTC().Unix()
}

func GetNowUnixMillo() int64 {
	return time.Now().UTC().UnixMilli()
}

func ZstdBytes(rawin []byte) []byte {
	enc, _ := zstd.NewWriter(nil)
	return enc.EncodeAll(rawin, nil)
}

func UnZstdBytes(zin []byte) (out []byte, err error) {
	dec, _ := zstd.NewReader(nil)
	out, err = dec.DecodeAll(zin, nil)
	if err != nil {
		PrintError("UnZstdBytes:DecodeAll", err)
		return nil, err
	}
	return out, nil
}

func GetEnv(k string, defaultVal string) string {
	ev := os.Getenv(k)
	if ev == "" {
		DebugWarn("GetEnv", "cannot find ENV var: [ ", k, " ], will use default value")
		return defaultVal
	}
	return ev
}

func MakeDirs(dpath string) error {
	_, err := os.Stat(dpath)
	if err != nil {
		DebugInfo("MakeDirs", dpath)
		err = os.MkdirAll(dpath, os.ModePerm)
		PrintError("MakeDirs:MkdirAll", err)
	}
	return nil
}

func ToUnixSlash(s string) string {
	// for windows
	return strings.ReplaceAll(s, "\\", "/")
}

func GetXxhash(b []byte) uint64 {
	return xxhash.Sum64(b)
}

func GetXxhashString(b []byte) string {
	return strconv.FormatUint(xxhash.Sum64(b), 10)
}

func SumBlake3(b []byte) []byte {
	h := blake3.New()
	h.Write(b)
	return []byte(fmt.Sprintf("%x", h.Sum(nil)))
}

func NewError(s string) error {
	return errors.New(s)
}

func GetPrimaryIP() string {
	addrs, err := net.InterfaceAddrs()
	PrintError("GetPrimaryIP", err)
	for _, addr := range addrs {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}

		if ip == nil || ip.IsLoopback() || ip.To4() == nil {
			continue
		}
		return ip.String()
	}
	return "0.0.0.0"
}

func Int2Str(n int) string {
	return strconv.Itoa(n)
}

func Int64ToString(n int64) string {
	return strconv.FormatInt(n, 10)
}

func Uint32ToString(n uint32) string {
	return fmt.Sprintf("%v", n)
}

func Uint64ToString(n uint64) string {
	return fmt.Sprintf("%v", n)
}

func Str2Int(n string) int {
	s, err := strconv.Atoi(n)
	if err != nil {
		return 0
	}
	return s
}

func Str2Int64(n string) int64 {
	s, err := strconv.ParseInt(n, 10, 64)
	if err != nil {
		return 0
	}
	return s
}

func Str2Uint64(n string) uint64 {
	s, err := strconv.ParseUint(n, 10, 64)
	if err != nil {
		return 0
	}
	return s
}

func IsAnyEmpty(args ...string) bool {
	for _, arg := range args {
		if arg == "" {
			return true
		}
	}
	return false
}

func IsAnyNil(args ...[]byte) bool {
	for _, arg := range args {
		if arg == nil {
			return true
		}
	}
	return false
}

func MapKeyOrdered(maps []map[string]int) []map[string]int {
	keySet := make(map[string]struct{})
	for _, m := range maps {
		for k := range m {
			keySet[k] = struct{}{}
		}
	}

	keys := make([]string, 0, len(keySet))
	for k := range keySet {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var result []map[string]int
	for _, k := range keys {
		for _, m := range maps {
			if val, exists := m[k]; exists {
				result = append(result, map[string]int{k: val})
			}
		}
	}

	return result
}

func ChModDir(dpath string, perm fs.FileMode) error {
	if err := os.Chmod(dpath, perm); err != nil {
		PrintError("ChModDir", err)
		return err
	}
	return nil
}

func Map2JSON(m map[string]string) []byte {
	bf := bytes.NewBuffer([]byte{})
	enc := json.NewEncoder(bf)
	enc.SetEscapeHTML(false)
	err := enc.Encode(m)
	if err != nil {
		PrintError("Map2JSON", err)
		return nil
	}
	return []byte(bf.String())
}

func JSON2Map(b []byte, m map[string]string) error {
	err := json.Unmarshal(b, &m)
	if err != nil {
		PrintError("JSON2Map", err)
		return err
	}
	return nil
}
