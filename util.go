package sqlconf

import (
	//"io"
	"io/ioutil"
	//"path/filepath"

	//"net"

	//"log"
	"net/http"
	"regexp"

	//"time"

	//"net/url"
	"os"
	//"strconv"
	"strings"

	"go.uber.org/zap"
)

func SliceUnique(sl []string) (u []string) {
	var m map[string]int = make(map[string]int, 16)
	for _, v := range sl {
		if v == "" {
			continue
		}
		m[v] = 0
	}
	for k, _ := range m {
		u = append(u, k)
	}
	return u
}

func StringToSlice(s string) (sl []string) {
	s = strings.ReplaceAll(s, ",", ";")
	s = strings.Trim(s, ";")
	ss := strings.Split(s, ";")

	for _, v := range ss {
		if v == "" {
			continue
		}
		sl = append(sl, v)
	}

	return sl
}

func SliceToString(l []string) (s string) {
	var ss []string
	for _, v := range l {
		if v == "" {
			continue
		}
		ss = append(ss, v)
	}

	return strings.Join(ss, ";")
}

func MakeDirs(s string) error {
	_, err := os.Stat(s)
	if err != nil {
		err := os.MkdirAll(s, os.ModePerm)
		if err != nil {
			zapLogger.Error("MakeDirs", zap.Error(err))
		} else {
			os.Chmod(s, os.ModePerm)
		}
	}
	return nil
}

func GetEnv(s string, d string) string {
	if os.Getenv(s) == "" {
		zapLogger.Warn("GetEnv:empty,use default instead", zap.String("env", s), zap.String("default", d))
		return d
	}
	return os.Getenv(s)
}

func GetURLContent(URL string) (cnt string, err error) {
	resp, err := http.Get(URL)
	if err != nil {
		zapLogger.Error("GetURLContent", zap.Error(err))
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		zapLogger.Error("GetURLContent", zap.Error(err))
		return "", err
	}

	return string(body), nil
}

func Filepathify(fp string) string {
	var replacement string = "_"

	reControlCharsRegex := regexp.MustCompile("[\u0000-\u001f\u0080-\u009f]")

	reRelativePathRegex := regexp.MustCompile(`^\.+`)

	filenameReservedRegex := regexp.MustCompile(`[<>:"\\|?*\x00-\x1F]`)
	filenameReservedWindowsNamesRegex := regexp.MustCompile(`(?i)^(con|prn|aux|nul|com[0-9]|lpt[0-9])$`)

	// reserved word
	fp = filenameReservedRegex.ReplaceAllString(fp, replacement)

	// continue
	fp = reControlCharsRegex.ReplaceAllString(fp, replacement)
	fp = reRelativePathRegex.ReplaceAllString(fp, replacement)
	fp = filenameReservedWindowsNamesRegex.ReplaceAllString(fp, replacement)
	return fp
}
