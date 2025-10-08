package cmd

import (
	"path/filepath"
)

func BeforeStart() error {
	DataDir = filepath.ToSlash(GetEnv("zstdb_data", "data/zstdfs"))
	if AltDataDir != "" {
		DataDir = AltDataDir
	}

	DebugInfo("BeforeStart: DataDir", DataDir)
	MakeDirs(DataDir)

	if MaxUploadSizeMB <= 0 {
		MaxUploadSizeMB = 16
	}

	if MaxUploadSizeMB > 1024 {
		MaxUploadSizeMB = 1024
	}

	MaxUploadSize = MaxUploadSizeMB << 20

	DebugInfo("MaxUploadSizeMB", MaxUploadSizeMB)

	return nil
}

func BeforeGrpcStart() error {
	//
	bgrdb = badgerConnect()
	DebugInfo("Max Version", bgrdb.MaxVersion())

	return nil
}
