
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOPRIVATE=codeup.aliyun.com

go mod tidy -compat="1.20"

:rem go get -u github.com/golang/protobuf
go get -u google.golang.org/protobuf
go get -u google.golang.org/grpc v1.52.3
go install google.golang.org/protobuf/cmd/protoc-gen-go
go get -u google.golang.org/grpc/health
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc
go get -u github.com/go-sql-driver/mysql 
go get -u github.com/jmoiron/sqlx
go get -u github.com/lestrrat-go/file-rotatelogs
go get -u github.com/sirupsen/logrus
go get -u github.com/go-redis/redis
go get -u github.com/gin-gonic/gin
go get -u github.com/coreos/etcd/mvcc/mvccpb
go get -u go.etcd.io/etcd/client/v3
go get -u github.com/fatih/structs
go get -u github.com/mitchellh/mapstructure
go get -u github.com/antlinker/go-dirtyfilter
go get -u github.com/aliyun/aliyun-oss-go-sdk/oss
go get -u github.com/gorilla/websocket
:rem go get -u github.com/fatih/color
:rem go get -u code.google.com/p/go.net/websocket
go get -u golang.org/x/time/rate

go get -u oversea-git.hotdogeth.com/poker/slots/svrframe

:rem http://127.0.0.1:2379/version