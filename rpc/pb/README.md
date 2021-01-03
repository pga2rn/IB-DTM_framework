# gRPC & protobuf code generation guide

1. generate grpc code

```shell
$ protoc --proto_path=$GOPATH/src -I.  --go_out . --go_opt paths=source_relative    --go-grpc_out . --go-grpc_opt paths=source_relative ./pb/framework.proto
```

2. generate grpc-gw

```shell
protoc --proto_path=$GOPATH/src -I . --grpc-gateway_out . --grpc-gateway_opt logtostderr=true --grpc-gateway_opt paths=source_relative      ./pb/framework.proto
```





