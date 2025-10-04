protoc --go_out=./pbs/ --go_opt=paths=source_relative --go-grpc_out=./pbs/ --go-grpc_opt=paths=source_relative badgerItem.proto

python -m grpc_tools.protoc -I . --python_out=./py/ --grpc_python_out=./py/ badgerItem.proto