package cmd

func BeforeGrpcStart() error {
	DataDir = GetEnv("zstdb_data", "data/zstdfs")
	DebugInfo("BeforeStart: DataDir", DataDir)
	MakeDirs(DataDir)
	if MaxUploadSizeMB <= 0 {
		MaxUploadSizeMB = 16
	}

	MaxUploadSize = MaxUploadSizeMB << 20

	bgrdb = badgerConnect()

	return nil
}
