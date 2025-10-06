CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o dist/macos_arm/zstdb -ldflags "-w -s" main.go
zip dist/macos_arm/zstdb_macos_arm.zip dist/macos_arm/zstdb

CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o dist/macos_intel/zstdb -ldflags "-w -s" main.go
zip dist/macos_intel/zstdb_macos_intel.zip dist/macos_intel/zstdb


CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/linux_amd64/zstdb -ldflags "-w -s" main.go
zip dist/linux_amd64/zstdb_linux_amd64.zip dist/linux_amd64/zstdb


CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -o dist/windows_amd64/zstdb.exe -ldflags "-w -s" main_windows.go
zip dist/windows_amd64/zstdb_windows_amd64.zip dist/windows_amd64/zstdb.exe
