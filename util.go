package sqlconf

import (
	"io"
	"io/ioutil"

	//"net"

	//"log"
	"net/http"
	"regexp"
	"time"

	//"net/url"
	"os"
	"strconv"
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

func DownloadFile(URL string, localPath string, isOverwrite bool, barTitle string) error {
	timeStart := time.Now().Unix()

	fi, err := os.Stat(localPath)

	if err == nil {
		if isOverwrite == true {
			if err = os.Remove(localPath); err != nil {
				zapLogger.Error("DownloadFile: error(os-remove)", zap.String("cannot delete file", localPath), zap.Error(err))
			}
		} else {
			zapLogger.Info("DownloadFile",
				zap.String("action", "SKIP"),
				zap.Bool("is-overwrite", isOverwrite),
				zap.Int64("size", fi.Size()),
				zap.Time("last-modified", fi.ModTime()),
				zap.String("localPath", localPath),
			)
			return nil
		}
	}

	req, err := http.NewRequest("GET", URL, nil)
	req.Header.Add("Accept-Encoding", "identity")
	req.Close = true

	if err != nil {
		zapLogger.Error("DownloadFile:error(http-NewRequest)", zap.Error(err))
		return err
	}
	resp, err := http.DefaultClient.Do(req)

	//resp, err := http.Get(URL)

	if err != nil {
		zapLogger.Error("DownloadFile:error(http-get)", zap.Error(err))
		return err
	} else {
		zapLogger.Info("Status",
			zap.String("proto", resp.Proto),
			zap.Int("status-code", resp.StatusCode))
	}
	defer resp.Body.Close()

	localPathTempName := strings.Join([]string{localPath, "downloading"}, ".")
	fileTemp, err := os.Create(localPathTempName)
	if err != nil {
		zapLogger.Error("DownloadFile:error(os-create)", zap.String("cannot create file", localPathTempName), zap.Error(err))
		return err
	}

	defer fileTemp.Close()

	var contentLength int64 = -1
	if resp.ContentLength > 0 {
		contentLength = resp.ContentLength
	}

	bar := Config.SetBar(contentLength, barTitle).Bar
	_, err = io.Copy(io.MultiWriter(fileTemp, bar), resp.Body)
	bar.Finish()

	if err != nil {
		zapLogger.Error("DownloadFile:error(io-copy)", zap.Error(err))
		return err
	}

	fileTemp.Close()

	err = os.Rename(localPathTempName, localPath)
	if err != nil {
		zapLogger.Error("DownloadFile:error(os-rename)", zap.Error(err))
		return err
	}

	timeStop := time.Now().Unix()

	zapLogger.Info("DownloadFile:ok",
		zap.String("proto", resp.Proto),
		zap.Int64("content-length", resp.ContentLength),
		zap.String("duration", strconv.FormatInt(timeStop-timeStart, 10)))

	return nil
}
