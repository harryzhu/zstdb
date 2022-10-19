package sqlconf

import (
	"io/ioutil"
	//"log"
	"net/http"

	//"time"

	//"net/url"
	"os"
	"strings"

	"go.uber.org/zap"
)

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
