module zstdb

go 1.25.1

require (
	github.com/cespare/xxhash/v2 v2.3.0
	github.com/dgraph-io/badger/v4 v4.8.0
	github.com/klauspost/compress v1.18.0
	github.com/spf13/cobra v1.10.1
	github.com/zeebo/blake3 v0.2.4
	google.golang.org/grpc v1.75.1
)

replace zstdb/pbs => ./proto/pbs

require (
	github.com/dgraph-io/ristretto/v2 v2.2.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/flatbuffers v25.2.10+incompatible // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/klauspost/cpuid/v2 v2.0.12 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel v1.37.0 // indirect
	go.opentelemetry.io/otel/metric v1.37.0 // indirect
	go.opentelemetry.io/otel/trace v1.37.0 // indirect
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/text v0.26.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250707201910-8d1bb00bc6a7 // indirect
	google.golang.org/protobuf v1.36.9 // indirect
	zstdb/pbs v0.0.0-00010101000000-000000000000 // indirect
)
