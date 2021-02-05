# mkit

golang 项目初始化工具


### 依赖
```bash
go get -u github.com/google/wire/cmd/wire
go get -u github.com/gogo/protobuf/proto
go get -u github.com/gogo/protobuf/protoc-gen-gogo
go get -u github.com/gogo/protobuf/gogoproto
go get -u github.com/golang/protobuf/{proto,protoc-gen-go}
go get -u google.golang.org/grpc
````


### 安装&使用
```bash
go get github.com/pescaria/mkit
mkit init yourAppName
```

$GOPATH/bin 需要添加到 $PATH.

### 构建&运行
```bash
make
bin/yourAppName -config=./configs/config.toml
```
