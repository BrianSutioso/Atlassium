module BrunoCoin

go 1.16

require (
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/phayes/freeport v0.0.0-20180830031419-95f893ade6f2
	go.uber.org/atomic v1.7.0
	golang.org/x/net v0.0.0-20190311183353-d8887717615a
	golang.org/x/sys v0.0.0-20210309074719-68d13333faf2 // indirect
	google.golang.org/grpc v1.37.0
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.1.0 // indirect
	google.golang.org/protobuf v1.26.0
)

replace BrunoCoin => ./
