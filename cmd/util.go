package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/klauspost/compress/zstd"
	"github.com/zeebo/blake3"
)

func GetNowUnix() int64 {
	return time.Now().UTC().Unix()
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
