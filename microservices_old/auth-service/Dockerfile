FROM golang:1.24.2 AS build

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

WORKDIR /build/auth-service

COPY auth-service/api /build/auth-service/api

WORKDIR /build

RUN go generate ./**/*api

COPY common/ /build/common/

COPY auth-service/ /build/auth-service/

RUN CGO_ENABLED=0 GOOS=linux go build \
      -ldflags="-w -s" \
      -installsuffix 'static' \
	  -o /auth-service /build/auth-service/main.go

FROM scratch

COPY --from=build /auth-service ./

EXPOSE 8080

ENTRYPOINT ["./auth-service"]
