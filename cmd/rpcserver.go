package cmd

import (
	"context"
	"fmt"
	"net"
	"path/filepath"
	"strings"

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
		Ver64:   0,
		Sum64:   0,
	}
	if in.Key != nil {
		val, ver := badgerGet(in.Key)
		if val != nil {
			resp.Errcode = 0
			resp.Status = []byte("ok")
			resp.Key = in.Key
			resp.Data = val
			resp.Ver64 = ver
			resp.Sum64 = GetXxhash(val)
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
		Key:     in.Key,
		Data:    nil,
		Ver64:   0,
		Sum64:   0,
	}
	if IsDisableSet == true {
		resp.Errcode = 501
		resp.Status = []byte("server disabled the set action")
		resp.Key = nil
		resp.Data = nil
		return resp, nil
	}

	if in.Data != nil {
		sum64 := GetXxhash(in.Data)
		if in.Sum64 != sum64 {
			resp.Errcode = 501
			resp.Status = []byte("data sum64 does not match")
			return resp, nil
		}

		k := badgerSave(in.Key, in.Data)
		if k != nil {
			resp.Key = k
			resp.Status = []byte("ok")
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
		Ver64:   0,
		Sum64:   0,
	}

	if IsDisableDelete == true {
		resp.Errcode = 501
		resp.Status = []byte("server disabled the delete action")
		resp.Key = nil
		resp.Data = nil
		return resp, nil
	}
	if in.Key != nil {
		err := badgerDelete(in.Key)
		if err != nil {
			resp.Errcode = 500
			resp.Status = []byte(err.Error())
			resp.Key = nil
			resp.Data = nil
		} else {
			resp.Status = []byte("ok")
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
		Ver64:   0,
		Sum64:   0,
	}
	if in.Key != nil {
		verNum := badgerExists(in.Key)

		if verNum == 0 {
			resp.Errcode = 404
			resp.Status = []byte("Not Found")
			resp.Ver64 = 0
		} else {
			resp.Errcode = 0
			resp.Status = []byte("ok")
			resp.Ver64 = verNum
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

func (s *server) Status(_ context.Context, in *pb.Item) (*pb.ItemReply, error) {
	resp := &pb.ItemReply{
		Errcode: 0,
		Status:  nil,
		Key:     in.Key,
		Data:    nil,
		Ver64:   0,
		Sum64:   0,
	}
	if in.Key != nil {
		inKey := strings.ToLower(string(in.Key))

		if inKey == "stats" {
			resp.Key = []byte("stats")

			var keyCount uint32 = 0
			tinfo := bgrdb.Tables()

			stats := make(map[string]string)
			for _, info := range tinfo {
				keyCount += info.KeyCount
			}
			lsm_size, vlog_size := bgrdb.Size()
			stats["max_version"] = Uint64ToString(bgrdb.MaxVersion())
			stats["key_count"] = Uint32ToString(keyCount)
			stats["lsm_size"] = Int64ToString(lsm_size)
			stats["vlog_size"] = Int64ToString(vlog_size)

			resp.Data = Map2JSON(stats)
		}

		if inKey == "backup" || inKey == "restore" {
			resp.Key = []byte(inKey)
			inData := in.Data
			m := make(map[string]string)
			err := JSON2Map(inData, m)
			if err != nil {
				PrintError(inKey, err)
				resp.Errcode = 500
				resp.Status = []byte(err.Error())
				return resp, nil
			}

			fpath := ""
			var fsince uint64 = 0
			for k, v := range m {
				if k == "path" {
					fpath = v
				}
				if k == "since" {
					fsince = Str2Uint64(v)
				}
			}

			if fpath == "" {
				resp.Errcode = 500
				resp.Status = []byte("path or since is invalid")
				return resp, nil
			}

			maxVer := bgrdb.MaxVersion()
			fdir := filepath.Dir(fpath)
			fname := strings.Join([]string{filepath.Base(fpath), "_", fmt.Sprintf("[%v-%v]", fsince, maxVer), ".zstdb.bak"}, "")
			err = MakeDirs(fdir)
			if err != nil {
				resp.Errcode = 500
				resp.Status = []byte(err.Error())
				return resp, nil
			}

			ftarget := filepath.Join(fdir, fname)

			if inKey == "backup" {
				badgerBackup(ftarget, fsince)
				m["target"] = ftarget
			}

			if inKey == "restore" {
				badgerRestore(fpath)
			}

			DebugInfo(inKey, "complete")

			resp.Data = Map2JSON(m)
		}

	}

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

	primaryIP := GetPrimaryIP()

	rpcServer = grpc.NewServer(opts...)
	pb.RegisterBadgerServer(rpcServer, &server{})
	DebugInfo("StartGrpcServer", "GRPC ADDRESS: ", addr)
	DebugInfo("StartGrpcServer", "GRPC(remote): ", primaryIP, ":", Port)
	DebugInfo("StartGrpcServer", "GRPC(local): ", "127.0.0.1:", Port)
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
