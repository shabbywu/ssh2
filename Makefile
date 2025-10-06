.DEFAULT_GOAL:=help

ifeq ($(shell go env GOOS),windows)
BINARY_SUFFIX := .exe
else
BINARY_SUFFIX :=
endif


##@ Build Binary
.PHONY: build
build: ## 构建
	go build -o bin/ssh2${BINARY_SUFFIX} -trimpath .


.PHONY: help
help:
	@awk 'BEGIN {FS = ":.*##"; printf "Usage: make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)