module go-corntab

go 1.15

replace (
	github.com/mongodb/mongo-go-driver v1.4.4 => go.mongodb.org/mongo-driver v1.4.4
	go.mongodb.org/mongo-driver v1.4.4 => github.com/mongodb/mongo-go-driver v1.4.4
)

require (
	github.com/coreos/etcd v3.3.25+incompatible
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/google/uuid v1.1.2 // indirect
	github.com/gorhill/cronexpr v0.0.0-20180427100037-88b0669f7d75
	github.com/mongodb/mongo-go-driver v1.4.4
	github.com/pkg/errors v0.9.1
	go.mongodb.org/mongo-driver v1.4.4
	go.uber.org/zap v1.16.0 // indirect
	google.golang.org/genproto v0.0.0-20201214200347-8c77b98c765d // indirect
	google.golang.org/grpc v1.34.0 // indirect
)

replace google.golang.org/grpc => google.golang.org/grpc v1.26.0

