.PHONY: build exec-d exec

build:
	go generate ./internal/config
	go build -v -o ./build/awww-gui ./cmd/awww-gui/main.go

exec-d:
	./build/awww-gui -d --log-level trace

exec:
	./build/awww-gui --log-level trace