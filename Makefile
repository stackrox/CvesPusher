export GO111MODULE := on

TAG=$(shell git describe --tags)
.PHONY: tag
tag:
	@echo $(TAG)

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
	docker build -t us.gcr.io/stackrox-hub/cve-uploader:$(TAG) image

.PHONY: push
push: image
	docker push us.gcr.io/stackrox-hub/cve-uploader:$(TAG) | cat
