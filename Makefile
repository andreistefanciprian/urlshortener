proto-url-gen:
	@protoc \
	--go_out=paths=source_relative:. \
	--go-grpc_out=paths=source_relative:. \
	./url-gen/proto/url-gen.proto

	@protoc \
	--go_out=paths=source_relative:./api-gateway/ \
	--go-grpc_out=paths=source_relative:./api-gateway/ \
	./url-gen/proto/url-gen.proto