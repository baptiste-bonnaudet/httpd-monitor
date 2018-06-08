



FROM golang:1.10-alpine

# build stage
FROM golang:alpine AS build-env

ADD ./src/ /go/src

RUN set -xe; \
  apk --no-cache add curl git; \
  curl https://glide.sh/get | sh;  \
  cd /go/src/app; \
  glide install; \
  env; \
  go build -o goapp;

# final stage
FROM alpine
WORKDIR /app
COPY --from=build-env /go/src/app/goapp /app/
ENTRYPOINT ./goapp
