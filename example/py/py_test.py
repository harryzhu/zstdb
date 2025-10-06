import hashlib
import time
import json
import grpc
import xxhash
import badgerItem_pb2
import badgerItem_pb2_grpc

rpc_addr = '192.168.0.113:8282'
#rpc_addr = '192.168.0.108:8282'
#rpc_addr = '127.0.0.1:8282'

max_msg_size = 1024*1024*1024
rpc_opt = (('grpc.max_send_message_length', max_msg_size),('grpc.max_receive_message_length', max_msg_size))

#with open("th.webp","rb")as fr:
with open("1.mp4","rb")as fr:
	fdata = fr.read()

def xxhashbyte(b):
	if b is None or len(b) == 0:
		return ""
	return xxhash.xxh64(b).intdigest()

def flist():
  with grpc.insecure_channel(target=rpc_addr, options=rpc_opt) as channel:
    stub = badgerItem_pb2_grpc.BadgerStub(channel)
    response = stub.List(badgerItem_pb2.ListFilter(prefix="",pagenum=1))
  print(response.keys)
  print(f'flist: {len(response.keys)}')
  print("-"*10)
 
def fset(k,v):
  with grpc.insecure_channel(target=rpc_addr, options=rpc_opt) as channel:
    stub = badgerItem_pb2_grpc.BadgerStub(channel)
    response = stub.Set(badgerItem_pb2.Item(key=k, data=v, sum64=xxhashbyte(v)))
    return response

def fsetKV(k,v):
  with grpc.insecure_channel(target=rpc_addr, options=rpc_opt) as channel:
    stub = badgerItem_pb2_grpc.BadgerStub(channel)
    response = stub.Set(badgerItem_pb2.Item(key=k, data=v, sum64=xxhashbyte(v)))
    return response


def fget(k):
  with grpc.insecure_channel(target=rpc_addr, options=rpc_opt) as channel:
    stub = badgerItem_pb2_grpc.BadgerStub(channel)
    response = stub.Get(badgerItem_pb2.Item(key=k.encode("utf-8")))
    print(f'fget({k}): data size: {len(response.data)} , xxhash:{xxhashbyte(response.data)}')
    return response

def fexists(k):
  with grpc.insecure_channel(target=rpc_addr, options=rpc_opt) as channel:
    stub = badgerItem_pb2_grpc.BadgerStub(channel)
    response = stub.Exists(badgerItem_pb2.Item(key=k.encode("utf-8")))
    return response

def fdelete(k):
  with grpc.insecure_channel(target=rpc_addr, options=rpc_opt) as channel:
    stub = badgerItem_pb2_grpc.BadgerStub(channel)
    response = stub.Delete(badgerItem_pb2.Item(key=k.encode("utf-8")))
    return response

def fstatus(k,v):
  with grpc.insecure_channel(target=rpc_addr, options=rpc_opt) as channel:
    stub = badgerItem_pb2_grpc.BadgerStub(channel)
    response = stub.Status(badgerItem_pb2.Item(key=k.encode("utf-8"), data=v.encode("utf-8")))
    return response
 
if __name__ == '__main__':
  test_key = None
  user_key = "my-key-01"
  #

  resp = fstatus("stats","")
  print(f'resp: fstatus: {resp}')

  # m = {"path": "/Users/harry/dev/apps/go10/zstdb/sdsf/sdf/sdfds", "since": "0"}
  # mstr = json.dumps(m)
  # resp = fstatus("backup",mstr)
  # print(f'resp: fstatus: {resp.data.decode("utf-8")}')

  # m = {"path": "/Users/harry/dev/apps/go10/zstdb/sdsf/sdf/sdfds_0.zstdb.bak", "since": "1"}
  # mstr = json.dumps(m)
  # resp = fstatus("restore",mstr)
  # print(f'resp: fstatus: {resp.data.decode("utf-8")}')

  
  
  flist()
  resp = fset("ff".encode("utf-8"), fdata)
  print(f'resp: fset: {resp}')
  print("-"*10)
  x = xxhashbyte(fdata)
  print(x)
  # # #
  # resp = fget('79e357782259ce7085eecb66777e1fd1918f817a329eb4d7d453c60fdcf94a54')
  # print(f'resp: fget: key: {resp.key.decode("utf-8")}, data size: {len(resp.data)}')
  # print("-"*10)
  # #
  # resp = fsetKV(user_key.encode("utf-8"),fdata)
  # print(f'resp: fsetKV: {resp}')
  # print("-"*10)
  # #
  print("#"*10)
  resp = fget("79e357782259ce7085eecb66777e1fd1918f817a329eb4d7d453c60fdcf94a54")
  print(f'resp: fget: key: {resp.key.decode("utf-8")}, sum64: {resp.sum64}, ver64: {resp.ver64}, data size: {len(resp.data)}')
  print("-"*10)
  # #
  resp = fexists("79e357782259ce7085eecb66777e1fd1918f817a329eb4d7d453c60fdcf94a54")
  print(f'resp: fexists: {resp}')
  print("-"*10)
  # #
  # resp = fdelete(user_key)
  # print(f'resp: fdelete: errcode:{resp.errcode}, key: {resp.key.decode("utf-8")}, data size: {len(resp.data)}')
  # print("-"*10)
  # # 
  # resp = fexists(user_key)
  # print(f'resp: fexists: key: {resp.key.decode("utf-8")}, data: {resp.data.decode("utf-8")}')
  # print("-"*10)
  # #
  resp = fget("4aeee52a218dea74ecbd857731ba317f811018ac7af13073e25640bcfecc9d2c")
  print(f'resp: fget: key: {resp.key.decode("utf-8")}, sum64: {resp.sum64}, ver64: {resp.ver64}, data size: {len(resp.data)}')
  print("-"*10)
  # #
  resp = fexists("4aeee52a218dea74ecbd857731ba317f811018ac7af13073e25640bcfecc9d2c")
  print(f'resp: fexists: {resp}')
  print("-"*10)

