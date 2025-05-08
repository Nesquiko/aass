# the root of this docker file is in the root of the project
FROM golang:1.24.2 AS build

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

WORKDIR /build/common

COPY common/server/api /build/common/server/api

RUN go generate ./...

WORKDIR /build/resource-service

COPY resource-service/api /build/resource-service/api

WORKDIR /build

RUN go generate ./...

COPY common/ /build/common/

COPY resource-service/ /build/resource-service/

RUN CGO_ENABLED=0 GOOS=linux go build \
      -ldflags="-w -s" \
      -installsuffix 'static' \
	  -o /resource-service /build/resource-service

FROM scratch

COPY --from=build /resource-service ./

EXPOSE 8080

ENTRYPOINT ["./resource-service"]
