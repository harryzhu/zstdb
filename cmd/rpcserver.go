package cmd

import (
	"context"
	"fmt"
	"net"

	pb "zstdb/pbs"

	"google.golang.org/grpc"
)

var rpcServer *grpc.Server

type server struct {
	pb.UnimplementedBadgerServer
}

func (s *server) Get(_ context.Context, in *pb.Item) (*pb.ItemReply, error) {
	resp := &pb.ItemReply{
		Errcode: 0,
		Status:  nil,
		Key:     in.Key,
		Data:    nil,
	}
	if in.Key != nil {
		v := badgerGet(in.Key)
		if v != nil {
			resp.Errcode = 0
			resp.Key = in.Key
			resp.Data = v
		} else {
			resp.Errcode = 500
			resp.Status = []byte("cannot get from bgrdb")
			resp.Data = nil
		}
	}

	return resp, nil
}

func (s *server) Set(_ context.Context, in *pb.Item) (*pb.ItemReply, error) {
	resp := &pb.ItemReply{
		Errcode: 0,
		Status:  nil,
		Key:     nil,
		Data:    nil,
	}
	if in.Data != nil {
		k := badgerSave(in.Key, in.Data)
		if k != nil {
			resp.Key = k
		} else {
			resp.Errcode = 500
			resp.Status = []byte("cannot save into bgrdb")
			resp.Key = nil
			resp.Data = nil
		}
	}

	return resp, nil
}

func (s *server) Delete(_ context.Context, in *pb.Item) (*pb.ItemReply, error) {
	resp := &pb.ItemReply{
		Errcode: 0,
		Status:  nil,
		Key:     in.Key,
		Data:    nil,
	}
	if in.Key != nil {
		err := badgerDelete(in.Key)
		if err != nil {
			resp.Errcode = 500
			resp.Status = []byte(err.Error())
		}

	}

	return resp, nil
}

func (s *server) Exists(_ context.Context, in *pb.Item) (*pb.ItemReply, error) {
	resp := &pb.ItemReply{
		Errcode: 0,
		Status:  nil,
		Key:     in.Key,
		Data:    nil,
	}
	if in.Key != nil {
		exist := badgerExists(in.Key)
		if exist == false {
			resp.Data = []byte("0")
		} else {
			resp.Data = []byte("1")
		}

	}

	return resp, nil
}

func (s *server) List(_ context.Context, in *pb.ListFilter) (*pb.ListFilterReply, error) {
	resp := &pb.ListFilterReply{}
	prefix := in.Prefix
	pagenum := int(in.Pagenum)

	resp.Keys = badgerList(prefix, pagenum)

	return resp, nil
}

func StartGrpcServer() {
	addr := fmt.Sprintf("%v:%v", Host, Port)
	lis, err := net.Listen("tcp", addr)
	FatalError("StartGrpcServer", err)
	//
	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(4096 * 1024 * 1024),
		grpc.MaxSendMsgSize(4096 * 1024 * 1024),
	}

	rpcServer = grpc.NewServer(opts...)
	pb.RegisterBadgerServer(rpcServer, &server{})
	DebugInfo("StartGrpcServer", "Local IP: ", GetPrimaryIP())
	DebugInfo("StartGrpcServer", "GRPC ADDRESS: ", addr)
	if err := rpcServer.Serve(lis); err != nil {
		FatalError("StartGrpcServer", err)
	}
}

func StopGrpcServer() {
	rpcServer.GracefulStop()
	bgrdb.Sync()
	bgrdb.Close()

	DebugInfo("StopGrpcServer", "stopping")

}
