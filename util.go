package sqlconf

import (
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
