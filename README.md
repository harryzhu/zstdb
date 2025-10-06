# zstdb

## Summary
一个使用 grpc 进行交互，供其他程序使用的 KV 键值后端存储服务，数据库采用 badger，值采用 zstd 压缩，相同文件仅存储一次，节省空间

## Prod
当前产品环境用于存储大量小文件（线上约 30万+ 文件大小约 1MB 的 .mp4 文件），相同文件仅存储一次。

## Usage
### 设置环境变量 `vim /etc/profile`
```Bash
    export zstdb_data="/Users/harry/data/zstdfs"

    # 指定数据存储路径
```

### 重新打开终端，运行
```Bash
./zstdb

# --debug 默认 false ： 是否显示各种调试信息
#
# --host 默认 false ： rpc 对外提供服务的 IP
# --port 默认 false ： rpc 对外提供服务的 端口 
#
# --max-upload-size-mb 默认 16 ： 值的最大长度，单位 MB
# --min-free-disk-space-mb 默认4096 ：设置最低磁盘可用空间，低于该值 zstdb 自动停止写入新数据，每10秒检测一次
#
# --allow-overwrite 默认 false ： 是否允许覆盖已经存在的值
# --allow-user-key 默认 false ： 是否允许用户自定义Key。默认不允许，目标是一个文件只存储一次，Key由系统自动生成
#
# --disable-delete 默认 false ： 禁用删除操作，数据库只允许添加数据，不允许删除数据
# --disable-set 默认 false ： 禁用写入操作，数据库不允许新添加数据，但可以删除数据

./zstdb >/dev/null 2>&1 &
# 后台运行

./zstdb &
# 后台运行
```

### 使用举例
#### Python
* 安装
```
pip install grpcio
pip install xxhash
```
* 写入数据格式:

```go
message Item {
  bytes key = 1;
  bytes data = 2;
  uint64 ver64 = 3;
  uint64 sum64 = 4;
}

// key: 当zstdb启动时，如果--allow-user-key=true，会用指定的该 key 存入数据，如果为 false，此处设置的key会被忽略
// data: 需要保存的数据
// ver64: 写入数据时，该值始终为0，查询返回时，为zstd内该数据的版本号，--allow-overwrite 设置为 true 时，该值会逐步递增，设置为 false 时，该值始终不变。
// sum64: 完整性校验，传入的数据，必须先在客户端采用 xxhash 得到哈希值，同时传入数据和这个哈希值，服务端接收数据后，会计算数据的 xxhash 值，
//        如果与客户端传入的 xxhash 值相同，才会认为接收的数据是完整的，才会写入数据库，客户端和服务端的 xxhash 值不相同时，数据不会被写入。
```

* 返回数据格式：
```go
message ItemReply {
  int32 errcode = 1;
  bytes status = 2;
  bytes key = 3;
  bytes data = 4;
  uint64 ver64 = 5;
  uint64 sum64 = 6;
}

```

* 支持方法： 
  * `Set`, 写入
  * `Get`, 读取
  * `Delete`, 删除
  * `Exists`, 检查数据是否存在，返回 0 表示不存在，返回其他数字表示：存在数据的当前版本号
  * `List`, 按指定前缀获取 Key 清单，分页，每次获取1000个Key。若前缀指定为空字符串，表示获取所有 ey
  * `Status`, 
    * `stats`, 获取简单统计数据 `max_version`, `key_count`, `lsm_size`, `vlog_size`
    * `backup`, 备份数据库，需要在 Data 字段提供 JSON 格式的 `path` 和 `since`, 值均为字符串。通过 since 的值可以增值备份
    * `restore`, 恢复数据库

```python

def fstatus(k,v):
  with grpc.insecure_channel(target=rpc_addr, options=rpc_opt) as channel:
    stub = badgerItem_pb2_grpc.BadgerStub(channel)
    response = stub.Status(badgerItem_pb2.Item(key=k.encode("utf-8"), data=v.encode("utf-8")))
    return response

# 备份数据库
m = {"path": "/Users/harry/zstdb/backup/20230101", "since": "0"}
# 注意： since 的值必须是 "0" (带有引号，即字符串)，不能是 0 （不带引号，即数字），
# since 的值是数据库内数据的版本号，"0" 表示全量备份，"232434" 表示仅备份版本为 "232434" 之后新增的数据，即增量备份
# path 为 zstdb 机器的本地路径, 文件名会添加后缀：_[since的值-备份时的最大版本号].zstdb.bak
mstr = json.dumps(m)
resp = fstatus("backup",mstr)
print(f'resp: fstatus: {resp.data.decode("utf-8")}')

# 恢复数据库
m = {"path": "/Users/harry/zstdb/backup/20230101_[0-132854].zstdb.bak"}
mstr = json.dumps(m)
resp = fstatus("restore",mstr)
print(f'resp: fstatus: {resp.data.decode("utf-8")}')
```

* 保存数据：


* 示例：

```python
import xxhash
import grpc
import badgerItem_pb2
import badgerItem_pb2_grpc

max_msg_size = 32*1024*1024

rpc_addr = '192.168.0.113:8282'
rpc_opt = (('grpc.max_send_message_length', max_msg_size),('grpc.max_receive_message_length', max_msg_size))

def xxhashbyte(b):
  if b is None or len(b) == 0:
    return None
  return xxhash.xxh64(b).intdigest()

def fset(k,v):
  with grpc.insecure_channel(target=rpc_addr, options=rpc_opt) as channel:
    stub = badgerItem_pb2_grpc.BadgerStub(channel)
    response = stub.Set(badgerItem_pb2.Item(key=k.encode("utf-8"), data=v, sum64=xxhashbyte(v)))
    return response

def fget(k):
  with grpc.insecure_channel(target=rpc_addr, options=rpc_opt) as channel:
    stub = badgerItem_pb2_grpc.BadgerStub(channel)
    response = stub.Get(badgerItem_pb2.Item(key=k.encode("utf-8")))
    return response


if __name__ == '__main__':
  with open("th.webp","rb")as fr:
    fdata = fr.read()

  resp = fset("my-test-key", fdata)
  # 3e18cf82e2b8416538a294f54a011359ba4b515d34e5a2195ac3231b6a9f3e17

  resp = fget("3e18cf82e2b8416538a294f54a011359ba4b515d34e5a2195ac3231b6a9f3e17")

  print(f'resp: errcode: {resp.errcode}')
  print(f'resp: status: {resp.status.decode("utf-8")}')
  print(f'resp: key: {resp.key.decode("utf-8")}')
  print(f'resp: data length: {len(resp.data)}')
  print(f'resp: ver64: {resp.ver64}')
  print(f'resp: sum64: {resp.sum64}')
  #
  # resp: errcode: 0
  # resp: status: ok
  # resp: key: 3e18cf82e2b8416538a294f54a011359ba4b515d34e5a2195ac3231b6a9f3e17
  # resp: data length: 244552
  # resp: ver64: 138701
  # resp: sum64: 16664423322944650346
  
  # ./zstdb 启动时，
  # 如果 --allow-user-key 设置为 true，那么 Key 就为 "my-test-key"
  # 如果 --allow-user-key 设置为 false， 那么 Key 就为 系统自动生成，而不是 user 设置的 "my-test-key"
  # 如果 --allow-overwrite 设置为 true， 那么新值会覆盖原值， 如果设置为 false，那么系统会忽略新值，不会更新原值
  #
  # 如果 --disable-set 设置为 true，那么文件不会被允许写入数据库，会直接返回501错误。
  #

```

