FROM golang:1.25.5-bookworm

RUN apt-get update \
    && apt-get install -y --no-install-recommends bash git ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download
RUN mkdir -p /go/build-cache /go/pkg/mod && chmod -R 0777 /go/build-cache /go/pkg/mod

CMD ["bash", "-c", "./scripts/test.sh"]
