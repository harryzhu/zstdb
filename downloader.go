package sqlconf

import (
	"io"
	//"io/ioutil"
	"path/filepath"

	//"net"

	//"log"
	"net/http"
	//"regexp"
	"time"

	//"net/url"
	"os"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

var (
	genFiles     []string
	maxAgeSecond int64
)

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
	} else {
		zapLogger.Info("plan",
			zap.String("localfile", localPathTempName),
			zap.String("url", URL))
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

func walkFunc(path string, finfo os.FileInfo, err error) error {
	var line string
	if finfo == nil {
		zapLogger.Warn("can't find:(" + path + ")")
		return nil
	}
	if path == "." || path == ".." {
		return nil
	}

	if finfo.IsDir() {
		return nil
	} else {
		ageSecond := ts_now - finfo.ModTime().Unix()
		if ageSecond < maxAgeSecond {
			line = strings.ReplaceAll(path, "\\", "/")
			genFiles = append(genFiles, line)
		} else {
			zapLogger.Info("walkFunc(skip)", zap.Int64("expired(age)", ageSecond), zap.String("file", path))
		}

		return nil
	}
}

func GenFileListByDir(dirPath, urlPrefix, outFile string, maxDays int64) error {
	if _, err := os.Stat(dirPath); err != nil {
		zapLogger.Error("GenListByDir", zap.Error(err))
		return err
	}

	err := filepath.Walk(dirPath, walkFunc)
	if err != nil {
		zapLogger.Error("GenListByDir-Walk", zap.Error(err))
		return err
	}

	if maxDays <= 0 {
		maxDays = 365 * 100
	}
	maxAgeSecond = 86400 * maxDays
	zapLogger.Info("GenFileListByDir", zap.Int64("max-age-second", maxAgeSecond))

	urlPrefix = strings.Trim(urlPrefix, "/")
	dirPathBackSlash := strings.ReplaceAll(dirPath, "\\", "/")
	dirPathBackSlash = strings.Trim(dirPathBackSlash, "/")

	var outLines []string
	outLines = append(outLines, strings.Join([]string{"### download-list", time.Now().Format("20060102T15:04:05Z07:00")}, "@"))
	for _, line := range genFiles {
		if filepath.Base(line) == filepath.Base(outFile) {
			continue
		}
		line = strings.Replace(line, dirPathBackSlash, urlPrefix, 1)
		outLines = append(outLines, line)
		zapLogger.Info(line)
	}
	outData := strings.Join(outLines, "\n")

	err = os.WriteFile(outFile, []byte(outData), os.ModePerm)
	if err != nil {
		return err
		zapLogger.Error("gen-file-list cannot be saved", zap.Error(err))
	} else {
		zapLogger.Info("gen-file-list was saved", zap.String("path", outFile))
	}

	return nil
}
