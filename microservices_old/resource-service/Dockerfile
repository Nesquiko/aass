FROM golang:1.24.2 AS build

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

WORKDIR /build/resource-service

COPY resource-service/api /build/resource-service/api

WORKDIR /build

RUN go generate ./**/*api

COPY common/ /build/common/

COPY resource-service/ /build/resource-service/

RUN CGO_ENABLED=0 GOOS=linux go build \
      -ldflags="-w -s" \
      -installsuffix 'static' \
	  -o /resource-service /build/resource-service/main.go

FROM scratch

COPY --from=build /resource-service ./

EXPOSE 8080

ENTRYPOINT ["./resource-service"]
