# zstdb

## Summary
一个使用 grpc 进行交互，供其他程序使用的 KV 键值数据库，值采用 zstd 压缩

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
```python
import grpc
import badgerItem_pb2
import badgerItem_pb2_grpc

rpcaddr = '192.168.0.113:8282'

max_msg_size = 32*1024*1024
rpc_opt = (('grpc.max_send_message_length', max_msg_size),('grpc.max_receive_message_length', max_msg_size))

def fset(k,v):
  with grpc.insecure_channel(target=rpc_addr, options=rpc_opt) as channel:
    stub = badgerItem_pb2_grpc.BadgerStub(channel)
    response = stub.Set(badgerItem_pb2.Item(key=k, data=v))
    return response

if __name__ == '__main__':
  with open("th.webp","rb")as fr:
    fdata = fr.read()

  resp = fset("my-test-key".encode("utf-8"), fdata)
  print(f'resp: fset: {resp}')
  # ./zstdb 启动时，
  # 如果 --allow-user-key 设置为 true，那么 Key 就为 "my-test-key"
  # 如果 --allow-user-key 设置为 false， 那么 Key 就为 系统自动生成，而不是 user 设置的 "my-test-key"
  # 如果 --allow-overwrite 设置为 true， 那么新值会覆盖原值， 如果设置为 false，那么系统会忽略新值，不会更新原值
  #
  # 如果 --disable-set 设置为 true，那么文件不会被允许写入数据库，会直接返回501错误。
  #

```

