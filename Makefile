export GO111MODULE := on

TAG=$(shell git describe --tags --abbrev=10 --dirty --long)
.PHONY: tag
tag:
	@echo $(TAG)

.PHONY: style
style:
	@go fmt ./...

.PHONY: test
test:
	@go test -v ./...

.PHONY: build
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
	go build -a \
		-ldflags "-s -w" \
		-o image/cve-uploader ./main

.PHONY: image
image: build
	docker build -t quay.io/stackrox-io/cve-uploader:$(TAG) image

.PHONY: push
push: image
	docker push quay.io/stackrox-io/cve-uploader:$(TAG) | cat

.PHONY: clean
clean:
	@rm image/cve-uploader || true
