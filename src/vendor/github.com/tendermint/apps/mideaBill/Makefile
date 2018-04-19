export PATH := $(GOPATH):$(PATH)
all:
	#protoc --js_out=import_style=commonjs,binary:. api.proto
	protoc --gofast_out=. api.proto
	#go test -v

