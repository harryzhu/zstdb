import hashlib
import os
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
rpc_admin_password = "123"
rpc_opt = (('grpc.max_send_message_length', max_msg_size),('grpc.max_receive_message_length', max_msg_size))


def xxhashbyte(b):
	if b is None or len(b) == 0:
		return ""
	return xxhash.xxh64(b).intdigest()

def flist():
  with grpc.insecure_channel(target=rpc_addr, options=rpc_opt) as channel:
    stub = badgerItem_pb2_grpc.BadgerStub(channel)
    response = stub.List(badgerItem_pb2.ListFilter(prefix="",pagenum=1))
  #print(response.keys)
  print(f'flist: {len(response.keys)}')
  print("-"*10)
  if len(response.keys) > 10:
    keys10 = response.keys[0:10]
  else:
    keys10 = response.keys[0:len(response.keys)-1]
  for k in keys10:
    print(f'{k}\n')
  return response
 
def fset(k,v):
  with grpc.insecure_channel(target=rpc_addr, options=rpc_opt) as channel:
    stub = badgerItem_pb2_grpc.BadgerStub(channel)
    response = stub.Set(badgerItem_pb2.Item(key=k.encode("utf-8"), data=v, sum64=xxhashbyte(v)))
    return response

def fsetKV(k,v):
  with grpc.insecure_channel(target=rpc_addr, options=rpc_opt) as channel:
    stub = badgerItem_pb2_grpc.BadgerStub(channel)
    response = stub.Set(badgerItem_pb2.Item(key=k.encode("utf-8"), data=v, sum64=xxhashbyte(v)))
    return response


def fget(k):
  with grpc.insecure_channel(target=rpc_addr, options=rpc_opt) as channel:
    stub = badgerItem_pb2_grpc.BadgerStub(channel)
    response = stub.Get(badgerItem_pb2.Item(key=k.encode("utf-8"),data='{}'.encode('utf-8'), sum64=xxhashbyte("{}")))
    #print(f'fget({k}): data size: {len(response.data)} , xxhash:{xxhashbyte(response.data)}')
    return response

def fexists(k):
  with grpc.insecure_channel(target=rpc_addr, options=rpc_opt) as channel:
    stub = badgerItem_pb2_grpc.BadgerStub(channel)
    response = stub.Exists(badgerItem_pb2.Item(key=k.encode("utf-8"),data='{}'.encode('utf-8'), sum64=xxhashbyte("{}")))
    return response

def fdelete(k):
  with grpc.insecure_channel(target=rpc_addr, options=rpc_opt) as channel:
    stub = badgerItem_pb2_grpc.BadgerStub(channel)
    response = stub.Delete(badgerItem_pb2.Item(key=k.encode("utf-8")))
    return response

def fadmin(k,v):
  with grpc.insecure_channel(target=rpc_addr, options=rpc_opt) as channel:
    stub = badgerItem_pb2_grpc.BadgerStub(channel)
    response = stub.Admin(badgerItem_pb2.Item(key=k.encode("utf-8"), 
      data=v.encode("utf-8"), 
      sum64=xxhashbyte(rpc_admin_password.encode("utf-8"))))
    return response
 
if __name__ == '__main__':
  test_key = None
  user_key = "my-key-01"
  #
  with open("th.webp","rb")as fr:
    fdata = fr.read()
  #
  test_data = {'name': "harry", 'age': 26, 'document': "oooook"}
  fdata = json.dumps(test_data).encode('utf-8')
  #
  resp = fset(user_key, fdata)
  print(f'resp: {resp}')
  fkey = resp.key.decode("utf-8")
  print(f'str_key: {fkey}')
  #
  resp = fget(fkey)
  print(f'resp: {resp}')
  resp_data = resp.data.decode('utf-8')
  print(f'resp_data: {resp_data}')
  j_data = json.loads(resp_data)
  print(f'json_data.name: {j_data["name"]}')
  #
  resp = fexists(fkey)
  print(f'resp: {resp}')
  resp_data = resp.data.decode('utf-8')
  print(f'resp_data: {resp_data}')
  j_data = json.loads(resp_data)
  print(f'json_data.exists: {j_data["exists"]}')
  #

  







