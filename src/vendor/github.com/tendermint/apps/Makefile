all:
	protoc --go_out=plugins=grpc:. mainapp.proto
	protoc --js_out=import_style=commonjs,binary:. mainapp.proto
	go install
