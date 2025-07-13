GIT_COMMIT := $(shell git rev-parse --short HEAD)
BUILD_DATE := $(shell date +%Y%m%d%H%M%S)
SERVICE_NAME := $(shell grep 'RUN_NAME=' build.sh | cut -d '=' -f 2)
TAG := $(SERVICE_NAME):$(GIT_COMMIT)-$(BUILD_DATE)

dependencies:
	go install github.com/cloudwego/kitex/tool/cmd/kitex@latest
	go install github.com/google/wire/cmd/wire@latest
	go install github.com/google/gnostic/cmd/protoc-gen-openapi@latest

kitex:
	kitex -I 3rd/idl/base api.proto