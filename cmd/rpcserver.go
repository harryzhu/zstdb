package cmd

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

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
		}

	}
	return resp, nil
}

func (s *server) Count(_ context.Context, in *pb.Item) (*pb.ItemReply, error) {
	resp := &pb.ItemReply{
		Errcode: 0,
		Status:  nil,
		Key:     in.Key,
		Data:    nil,
		Ver64:   0,
		Sum64:   0,
	}

	if in.Key != nil {
		numPrefix := badgerCount(string(in.Key))
		resp.Data = []byte(Uint64ToString(numPrefix))
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
			resp.Ver64 = verNum
		}

	}

	return resp, nil
}

func (s *server) List(_ context.Context, in *pb.ListFilter) (*pb.ListFilterReply, error) {
	resp := &pb.ListFilterReply{}
	prefix := in.Prefix
	pagenum := int(in.Pagenum)
	badgerSync()
	resp.Keys = badgerList(prefix, pagenum)

	return resp, nil
}

func (s *server) Admin(_ context.Context, in *pb.Item) (*pb.ItemReply, error) {
	resp := &pb.ItemReply{
		Errcode: 0,
		Status:  nil,
		Key:     in.Key,
		Data:    nil,
		Ver64:   0,
		Sum64:   0,
	}

	if in.Sum64 != GetXxhash([]byte(AdminPassword)) {
		resp.Errcode = 403
		resp.Status = []byte("incorrect  password")
		return resp, nil
	}

	badgerSync()

	if in.Key != nil {
		inKey := strings.ToLower(string(in.Key))
		resp.Key = []byte(inKey)

		if inKey == "stop" {
			go func() {
				time.Sleep(2 * time.Second)
				StopGrpcServer()
				time.Sleep(2 * time.Second)
				os.Exit(0)
			}()

			return resp, nil
		}

		if inKey == "gc" {
			err := bgrdb.RunValueLogGC(0.5)
			if err != nil {
				resp.Status = []byte(err.Error())
			}

			return resp, nil
		}

		if inKey == "sync" {
			err := badgerSync()
			if err != nil {
				resp.Status = []byte(err.Error())
			}

			return resp, nil
		}

		if inKey == "status" {
			var keyCount uint64 = 0
			stats := make(map[string]string)

			t1 := GetNowUnixMillo()
			keyCount = badgerCount("")
			tElapse := GetNowUnixMillo() - t1

			lsm_size, vlog_size := bgrdb.Size()
			stats["max_version"] = Uint64ToString(bgrdb.MaxVersion())
			stats["key_count"] = Uint64ToString(keyCount)
			stats["lsm_size"] = Int64ToString(lsm_size)
			stats["vlog_size"] = Int64ToString(vlog_size)
			stats["elapse_ms"] = Int64ToString(tElapse)

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

			fdir := filepath.Dir(fpath)
			err = MakeDirs(fdir)
			if err != nil {
				resp.Errcode = 500
				resp.Status = []byte(err.Error())
				return resp, nil
			}

			if inKey == "backup" {
				err := badgerBackup(fpath, fsince)
				if err == nil {
					doneFile := strings.Join([]string{fpath, "backup.done"}, ".")
					doneContent := ReadFile(doneFile)
					if doneContent != nil {
						m["target"] = string(doneContent)
					}
				} else {
					resp.Errcode = 500
					resp.Status = []byte(err.Error())
					m["target"] = ""
				}
			}

			if inKey == "restore" {
				err := badgerRestore(fpath)
				if err != nil {
					resp.Errcode = 500
					resp.Status = []byte(err.Error())
					m["target"] = "failed"
				} else {
					m["target"] = "ok"
				}
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
	DebugInfo("StopGrpcServer", "Stopping ...")
	ScheduleTask.Stop()
	rpcServer.GracefulStop()
	badgerSync()
	bgrdb.Close()

	if pidFile != "" {
		_, err := os.Stat(pidFile)
		if err == nil {
			RemoveFile(pidFile)
		}
	}

	if rpcFile != "" {
		_, err := os.Stat(rpcFile)
		if err == nil {
			RemoveFile(rpcFile)
		}
	}

	StopFileLogging()
	DebugInfo("StopGrpcServer", "Done")
}
