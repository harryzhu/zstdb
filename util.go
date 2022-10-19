package sqlconf

import (
	"log"
	"os"
	"strings"
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
			log.Fatal(err)
		} else {
			os.Chmod(s, os.ModePerm)
		}
	}
	return nil
}
