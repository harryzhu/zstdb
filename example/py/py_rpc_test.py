import hashlib
import os
import time
import json
import grpc
import xxhash
import badgerItem_pb2
import badgerItem_pb2_grpc

#rpc_addr = '192.168.0.113:8282'
rpc_addr = '192.168.0.108:8282'
#rpc_addr = '127.0.0.1:8282'

rpc_admin_password = "123"
max_msg_size = 1024*1024*1024
rpc_opt = (('grpc.max_send_message_length', max_msg_size),('grpc.max_receive_message_length', max_msg_size))

def xxhash_byte(b):
	if b is None or len(b) == 0:
		return 0
	return xxhash.xxh64(b).intdigest()

class zstdb():
  def __init__(self,key: str, data: bytes):
    #print(rpc_addr)
    self.item=badgerItem_pb2.Item(key = key.encode("utf-8"),data = data,ver64 = 0,sum64 = xxhash_byte(data))
    self.list_filter=badgerItem_pb2.ListFilter(prefix=key,pagenum=1)
    self.admin_item=badgerItem_pb2.Item(key = key.encode("utf-8"),data = data,ver64 = 0,sum64 = xxhash_byte(rpc_admin_password))
    self.response = ""
    self.rpc_cmd = ""
    channel = grpc.insecure_channel(target=rpc_addr, options=rpc_opt)
    self.stub = badgerItem_pb2_grpc.BadgerStub(channel)

  def set(self):
    self.rpc_cmd = "set"
    self.response = self.stub.Set(self.item)
    return self

  def get(self):
    self.rpc_cmd = "get"
    self.response = self.stub.Get(self.item)
    return self

  def exists(self):
    self.rpc_cmd = "exists"
    self.response = self.stub.Exists(self.item)
    return self

  def count(self):
    self.rpc_cmd = "count"
    self.response = self.stub.Count(self.item)
    return self

  def delete(self):
    self.rpc_cmd = "delete"
    self.response = self.stub.Delete(self.item)
    return self

  def list(self):
    self.rpc_cmd = "list"
    self.response = self.stub.List(self.list_filter)
    return self

  def ping(self):
    self.rpc_cmd = "ping"
    self.response = self.stub.Ping(self.item)
    return self

  def admin(self,cmd):
    self.rpc_cmd = f'admin: {cmd}'
    self.response = self.stub.Admin(self.admin_item)
    return self

  def print_item_reply(self, is_decode_data = False):
    if self.response is None:
      print("no response")
      return
    print(f'\n{"#"*10} {self.rpc_cmd} {"#"*10}')
    print(f'errcode: {self.response.errcode}')
    print(f'status: {self.response.status.decode("utf-8")}')
    print(f'key: {self.response.key.decode("utf-8")}')
    if is_decode_data:
      print(f'data: {self.response.data.decode("utf-8")}')
    else:
      print(f'data length: {len(self.response.data)}')
    print(f'ver64: {self.response.ver64}')
    print(f'sum64: {self.response.sum64}')

  def print_list_reply(self):
    if self.response is None:
      print("no response")
      return
    print(f'\n{"#"*10} {self.rpc_cmd} {"#"*10}')
    print(f'key: {self.response.keys}')

  def print_count_reply(self):
    if self.response is None:
      print("no response")
      return
    print(f'\n{"#"*10} {self.rpc_cmd} {"#"*10}')
    print(f'errcode: {self.response.errcode}')
    print(f'status: {self.response.status.decode("utf-8")}')
    print(f'key: {self.response.key}')
    print(f'data: {self.response.data.decode("utf-8")}')

  def print_admin_reply(self):
    if self.response is None:
      print("no response")
      return
    print(f'\n{"#"*10} {self.rpc_cmd} {"#"*10}')
    print(f'errcode: {self.response.errcode}')
    print(f'status: {self.response.status.decode("utf-8")}')
    print(f'key: {self.response.key}')
    print(f'data: {self.response.data.decode("utf-8")}')
    


 
if __name__ == '__main__':
  print("-"*50)
  with open("th.webp","rb")as fr:
    fdata = fr.read()

  z = zstdb("",b'')
  z.ping().print_item_reply(True)
  #
  z = zstdb("my-test-key",fdata)
  z.set().print_item_reply()
  #
  z = zstdb("3e18cf82e2b8416538a294f54a011359ba4b515d34e5a2195ac3231b6a9f3e17",b'')
  z.get().print_item_reply()
  #
  exists = {"mode": 0}
  byte_exists = json.dumps(exists).encode("utf-8")
  z = zstdb("3e18cf82e2b8416538a294f54a011359ba4b515d34e5a2195ac3231b6a9f3e17",byte_exists)
  z.exists().print_item_reply(True)
  #
  #z = zstdb("3e18cf82e2b8416538a294f54a011359ba4b515d34e5a2195ac3231b6a9f3e17",b'')
  #z.delete().print_item_reply()
  #
  exists = {"mode": 1}
  byte_exists = json.dumps(exists).encode("utf-8")
  print(byte_exists)
  z = zstdb("3e18cf82e2b8416538a294f54a011359ba4b515d34e5a2195ac3231b6a9f3e17",byte_exists)
  z.exists().print_item_reply(True)
  #
  # z = zstdb("my-test-key",fdata)
  # z.set().print_item_reply()
  # #
  z = zstdb("1234",b'')
  z.count().print_count_reply()
  # #
  z = zstdb("1234",b'')
  z.list().print_list_reply()
  # #
  z = zstdb("status",b'')
  z.admin("status").print_admin_reply()
  #
  #backup = {"path": "/Users/harry/backup/20230312", "since": "0"}
  # backup = {"path": "E:/backup/20230312", "since": "0"}
  # byte_backup = json.dumps(backup).encode("utf-8")
  # z = zstdb("backup",byte_backup)
  # z.admin("backup").print_admin_reply()
  #
  # restore = {"path": "E:/backup/20230312_[0_31].zstdb.bak"}
  # byte_restore = json.dumps(restore).encode("utf-8")
  # z = zstdb("restore",byte_restore)
  # z.admin("restore").print_admin_reply()











