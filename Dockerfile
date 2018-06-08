



FROM golang:1.10-alpine

# build stage
FROM golang:alpine AS build-env

ADD ./src /src

RUN set -xe; \
  apk --no-cache add curl git; \
  cd /src; \
  curl https://glide.sh/get | sh;  \
  glide install; \
  go build -o goapp;

# final stage
FROM alpine
WORKDIR /app
COPY --from=build-env /src/goapp /app/
ENTRYPOINT ./goapp
