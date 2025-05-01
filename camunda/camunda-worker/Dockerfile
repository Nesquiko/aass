FROM golang:1.24.2 AS build

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

WORKDIR /build/common

COPY common/server/api /build/common/server/api

RUN go generate ./...

WORKDIR /build/camunda-worker

COPY camunda-worker/resources-api /build/camunda-worker/resources-api

WORKDIR /build

RUN go generate ./...

COPY common/ /build/common/

COPY camunda-worker/ /build/camunda-worker/

RUN CGO_ENABLED=0 GOOS=linux go build \
      -ldflags="-w -s" \
      -installsuffix 'static' \
	  -o /camunda-worker /build/camunda-worker

FROM scratch

COPY --from=build /camunda-worker ./

EXPOSE 8080

ENTRYPOINT ["./camunda-worker"]

