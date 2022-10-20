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
	go build -ldflags "-s -w" \
		-o image/cve-uploader ./main

.PHONY: dry-run
dry-run: build
	image/cve-uploader -dry-run

.PHONY: image
image: build
	docker build -t quay.io/rhacs-eng/qa:cve-uploader-$(TAG) image

.PHONY: push
push: image
	docker push quay.io/rhacs-eng/qa:cve-uploader-$(TAG) | cat

.PHONY: clean
clean:
	@rm image/cve-uploader || true
