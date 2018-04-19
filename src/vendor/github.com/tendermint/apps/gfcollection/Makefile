all:
	protoc --go_out=plugins=grpc:. api.proto
	protoc --js_out=import_style=commonjs,binary:. api.proto
	go test -v
