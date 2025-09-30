proto-url-gen:
	@protoc \
	--go_out=paths=source_relative:. \
	--go-grpc_out=paths=source_relative:. \
	./url-gen/proto/url-gen.proto

	@protoc \
	--go_out=paths=source_relative:./api-gateway/ \
	--go-grpc_out=paths=source_relative:./api-gateway/ \
	./url-gen/proto/url-gen.proto

proto-url-read:
	@protoc \
	--go_out=paths=source_relative:. \
	--go-grpc_out=paths=source_relative:. \
	./url-read/proto/url-read.proto

	@protoc \
	--go_out=paths=source_relative:./api-gateway/ \
	--go-grpc_out=paths=source_relative:./api-gateway/ \
	./url-read/proto/url-read.proto

# Docker Compose commands
up:
	@echo "Starting URL Shortener services..."
	@docker compose up --build

down:
	@echo "Stopping URL Shortener services and cleaning up..."
	@sudo docker compose down --volumes --remove-orphans
	@echo "All services stopped and cleaned up!"

.PHONY: proto-url-gen proto-url-read up down