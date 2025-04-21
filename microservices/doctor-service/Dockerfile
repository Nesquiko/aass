FROM golang:1.24.2 AS build

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

WORKDIR /build/doctor-service

COPY doctor-service/api /build/doctor-service/api

WORKDIR /build

RUN go generate ./**/*api

COPY common/ /build/common/

COPY doctor-service/ /build/doctor-service/

RUN CGO_ENABLED=0 GOOS=linux go build \
      -ldflags="-w -s" \
      -installsuffix 'static' \
	  -o /doctor-service /build/doctor-service/main.go

FROM scratch

COPY --from=build /doctor-service ./

EXPOSE 8080

ENTRYPOINT ["./doctor-service"]
