FROM golang:1.24.2 AS build

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

WORKDIR /build/prescription-service

COPY prescription-service/api /build/prescription-service/api

WORKDIR /build

RUN go generate ./**/*api

COPY common/ /build/common/

COPY prescription-service/ /build/prescription-service/

RUN CGO_ENABLED=0 GOOS=linux go build \
      -ldflags="-w -s" \
      -installsuffix 'static' \
	  -o /prescription-service /build/prescription-service/main.go

FROM scratch

COPY --from=build /prescription-service ./

EXPOSE 8080

ENTRYPOINT ["./prescription-service"]
