VERSION 0.8
FROM golang:1.22.1
WORKDIR /workdir


codegen-client:
  RUN go install github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen@latest
  COPY spec.yaml .
  RUN oapi-codegen -package client spec.yaml > client.gen.go
  SAVE ARTIFACT client.gen.go AS LOCAL pkg/client/client.gen.go

# might consider implementing submodules instead
json-data:
  RUN mkdir data
  RUN curl https://raw.githubusercontent.com/helldivers-2/json/master/planets.json -o data/planets.json
  SAVE ARTIFACT data AS LOCAL data

build-exporter:
  FROM DOCKERFILE --build-arg COMMAND=exporter .
  SAVE IMAGE sigbilly/hde_exporter:latest
