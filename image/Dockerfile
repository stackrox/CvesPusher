FROM alpine:3.16

RUN apk add --no-cache ca-certificates

COPY cve-uploader /usr/bin/cve-uploader

ENTRYPOINT ["/usr/bin/cve-uploader"]
