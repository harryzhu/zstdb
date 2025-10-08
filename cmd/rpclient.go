package cmd

import (
	"context"
	"time"

	pb "zstdb/pbs"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	gConn   *grpc.ClientConn
	gClient pb.BadgerClient
)

func SetGrpcClient(rpcAddr string) {
	var err error
	gConn, err = grpc.NewClient(rpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	PrintError("SetClient", err)
	gClient = pb.NewBadgerClient(gConn)
}

func gcAdminStop() error {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*15)
	r, err := gClient.Admin(ctx, &pb.Item{Key: []byte("stop"), Sum64: GetXxhash([]byte(stopRpcAdminPassword))})
	if err != nil {
		PrintError("gcAdminStop", err)
		return err
	}
	DebugInfo("gcAdminStop", string(r.Status))
	return nil
}
