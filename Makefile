.PHONY: build exec-d exec

build:
	go build -v -o ./bin/awww-gui ./cmd/awww-gui/main.go

exec-d:
	./bin/awww-gui -d --log-level trace

exec:
	./bin/awww-gui --log-level trace