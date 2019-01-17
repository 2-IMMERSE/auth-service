FROM golang:alpine AS build

ARG CGO_ENABLED=1
ARG GOOS=linux
ARG GOARCH=amd64

WORKDIR /go/src/2-immerse/auth-service

ADD . /go/src/2-immerse/auth-service

RUN apk add --no-cache git
RUN go get -d -v ./...
RUN go build --ldflags="-s"

FROM alpine:3.6

COPY --from=build /go/src/2-immerse/auth-service/auth-service /auth-service
COPY schema /schema
COPY fixtures /fixtures

RUN apk add --no-cache ca-certificates \
    && chmod +x /auth-service

EXPOSE 8080

ENTRYPOINT ["/auth-service"]
