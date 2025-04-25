FROM golang:1.24.2 AS build

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

WORKDIR /build/condition-service

COPY condition-service/api /build/condition-service/api

WORKDIR /build

RUN go generate ./**/*api

COPY common/ /build/common/

COPY condition-service/ /build/condition-service/

RUN CGO_ENABLED=0 GOOS=linux go build \
      -ldflags="-w -s" \
      -installsuffix 'static' \
	  -o /condition-service /build/condition-service/main.go

FROM scratch

COPY --from=build /condition-service ./

EXPOSE 8080

ENTRYPOINT ["./condition-service"]
