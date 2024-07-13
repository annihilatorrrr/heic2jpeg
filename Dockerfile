FROM golang:1.22.5-alpine3.20 as builder
WORKDIR /heic2jpeg
RUN apk update && apk upgrade --available && sync && apk add --no-cache --virtual .build-deps build-base git gcc libc-dev libpng-dev jpeg-dev libwebp-dev libtiff-dev zlib-dev
COPY . .
RUN go build -ldflags="-w -s" .
FROM alpine:3.20.1
RUN apk update && apk upgrade --available && sync && apk add --no-cache --virtual .build-deps libpng-dev jpeg-dev libwebp-dev libtiff-dev zlib-dev
COPY --from=builder /heic2jpeg/heic2jpeg /heic2jpeg
ENTRYPOINT ["/heic2jpeg"]