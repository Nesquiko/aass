FROM golang:1.24.2 AS build

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

WORKDIR /build/patient-service

COPY patient-service/api /build/patient-service/api

WORKDIR /build

RUN go generate ./**/api

COPY common/ /build/common/

COPY patient-service/ /build/patient-service/

RUN CGO_ENABLED=0 GOOS=linux go build \
      -ldflags="-w -s" \
      -installsuffix 'static' \
	  -o /patient-service /build/patient-service/main.go

FROM scratch

COPY --from=build /patient-service ./

EXPOSE 8080

ENTRYPOINT ["./patient-service"]
