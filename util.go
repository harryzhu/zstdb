package sqlconf

import (
	"io"
	"io/ioutil"

	//"log"
	"net/http"

	"time"

	//"net/url"
	"os"
	"strconv"
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

func DownloadFile(URL string, localPath string, isOverwrite bool) error {
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

	resp, err := http.Get(URL)

	if err != nil {
		zapLogger.Error("DownloadFile:error(http-get)", zap.Error(err))
		return err
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

	bar := Config.SetBar(contentLength, "downloading").Bar
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
