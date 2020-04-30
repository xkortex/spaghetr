# Compile protobuf files in this directory for test example.

.PHONY: all proto test go clean

all: proto go

proto:
	python -m protosanity.compile_pbs -p grpc_go spaghetr/protos

go:
	go build -o ${GOPATH}/bin/spaghetr.server ./server/server.go
	go build -o ${GOPATH}/bin/spaghetr.client ./client/client.go

# have to cd elsewhere
test: proto go

clean:
	rm -f spaghetr/protos/*_pb2.py
	rm -f spaghetr/protos/*_pb2.pyi
	rm -f spaghetr/protos/*_pb2_grpc.py
	rm -f spaghetr/protos/*.pb.go
