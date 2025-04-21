FROM golang:1.24.2 AS build

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

WORKDIR /build/medical-history-service

COPY medical-history-service/api /build/medical-history-service/api

WORKDIR /build

RUN go generate ./**/*api

COPY common/ /build/common/

COPY medical-history-service/ /build/medical-history-service/

RUN CGO_ENABLED=0 GOOS=linux go build \
      -ldflags="-w -s" \
      -installsuffix 'static' \
	  -o /medical-history-service /build/medical-history-service/main.go

FROM scratch

COPY --from=build /medical-history-service ./

EXPOSE 8080

ENTRYPOINT ["./medical-history-service"]
