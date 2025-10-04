import hashlib
import grpc
import badgerItem_pb2
import badgerItem_pb2_grpc

rpcaddr = '192.168.0.113:8282'

with open("th.webp","rb")as fr:
	fdata = fr.read()

def md5byte(b):
	if b is None or len(b) == 0:
		return ""
	return hashlib.md5(b).hexdigest()

def flist():
  with grpc.insecure_channel(rpcaddr) as channel:
    stub = badgerItem_pb2_grpc.BadgerStub(channel)
    response = stub.List(badgerItem_pb2.ListFilter(prefix="",pagenum=1))
  print(response.keys)
  print(f'flist: {len(response.keys)}')
  print("-"*10)
 
def fset(k,v):
  with grpc.insecure_channel(rpcaddr) as channel:
    stub = badgerItem_pb2_grpc.BadgerStub(channel)
    response = stub.Set(badgerItem_pb2.Item(key=k, data=v))
    return response

def fsetKV(k,v):
  with grpc.insecure_channel(rpcaddr) as channel:
    stub = badgerItem_pb2_grpc.BadgerStub(channel)
    response = stub.Set(badgerItem_pb2.Item(key=k, data=v))
    return response


def fget(k):
  with grpc.insecure_channel(rpcaddr) as channel:
    stub = badgerItem_pb2_grpc.BadgerStub(channel)
    response = stub.Get(badgerItem_pb2.Item(key=k.encode("utf-8")))
    print(f'fget({k}): data size: {len(response.data)} , md5:{md5byte(response.data)}')
    return response

def fexists(k):
  with grpc.insecure_channel(rpcaddr) as channel:
    stub = badgerItem_pb2_grpc.BadgerStub(channel)
    response = stub.Exists(badgerItem_pb2.Item(key=k.encode("utf-8")))
    return response

def fdelete(k):
  with grpc.insecure_channel(rpcaddr) as channel:
    stub = badgerItem_pb2_grpc.BadgerStub(channel)
    response = stub.Delete(badgerItem_pb2.Item(key=k.encode("utf-8")))
    return response
 
if __name__ == '__main__':
  test_key = None
  user_key = "my-key-01"
  #
  flist()
  resp = fset("ff".encode("utf-8"), fdata)
  print(f'resp: fset: {resp}')
  print("-"*10)
  #
  resp = fget(resp.key.decode("utf-8"))
  print(f'resp: fget: key: {resp.key.decode("utf-8")}, data size: {len(resp.data)}')
  print("-"*10)
  #
  resp = fsetKV(user_key.encode("utf-8"),fdata)
  print(f'resp: fsetKV: {resp}')
  print("-"*10)
  #
  resp = fget(resp.key.decode("utf-8"))
  print(f'resp: fget: key: {resp.key.decode("utf-8")}, data size: {len(resp.data)}')
  print("-"*10)
  #
  resp = fexists(user_key)
  print(f'resp: fexists: key: {resp.key.decode("utf-8")}, data: {resp.data.decode("utf-8")}')
  print("-"*10)
  #
  resp = fdelete(user_key)
  print(f'resp: fdelete: key: {resp.key.decode("utf-8")}, data size: {len(resp.data)}')
  print("-"*10)
  # 
  resp = fexists(user_key)
  print(f'resp: fexists: key: {resp.key.decode("utf-8")}, data: {resp.data.decode("utf-8")}')
  print("-"*10)
  #


