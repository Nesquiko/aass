# the root of this docker file is in the root of the project
FROM golang:1.24.2 AS build

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

WORKDIR /build/common

COPY common/server/api /build/common/server/api

RUN go generate ./...

WORKDIR /build/medical-service

COPY medical-service/api /build/medical-service/api

WORKDIR /build

RUN go generate ./...

COPY common/ /build/common/

COPY medical-service/ /build/medical-service/

RUN CGO_ENABLED=0 GOOS=linux go build \
      -ldflags="-w -s" \
      -installsuffix 'static' \
	  -o /medical-service /build/medical-service

FROM scratch

COPY --from=build /medical-service ./

EXPOSE 8080

ENTRYPOINT ["./medical-service"]
