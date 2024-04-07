FROM golang:1.22.1
WORKDIR /go/src/app
COPY go.sum go.mod ./
RUN go mod download
COPY . .
RUN go build \
  -o /go/bin/app \
  -ldflags "-X github.com/prometheus/common/version.Version=0.0.1"
FROM gcr.io/distroless/base-debian12
COPY data /data
COPY --from=0 /go/bin/app /app
ENTRYPOINT ["/app"]
