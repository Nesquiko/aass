FROM golang:1.24.2 AS build

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

WORKDIR /build/appointment-service

COPY appointment-service/api /build/appointment-service/api

WORKDIR /build

RUN go generate ./**/*api

COPY common/ /build/common/

COPY appointment-service/ /build/appointment-service/

RUN CGO_ENABLED=0 GOOS=linux go build \
      -ldflags="-w -s" \
      -installsuffix 'static' \
	  -o /appointment-service /build/appointment-service/main.go

FROM scratch

COPY --from=build /appointment-service ./

EXPOSE 8080

ENTRYPOINT ["./appointment-service"]
