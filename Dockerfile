FROM golang:1.22.5-alpine3.20 as builder
WORKDIR /heic2jpeg
RUN apk update && apk upgrade --available && sync && apk add --no-cache build-base git gcc libc-dev libpng-dev jpeg-dev libwebp-dev tiff-dev zlib-dev
COPY . .
RUN curl -sSL https://github.com/libvips/libvips/releases/download/v8.15.2/vips-8.15.2.tar.xz | tar xJ && cd vips-8.15.2 && ./configure && make && make install && ldconfig
RUN rm -rf /var/cache/apk/* /vips-*
RUN go build -ldflags="-w -s" .
FROM alpine:3.20.1
COPY --from=builder /usr/local/bin/vips /usr/local/bin/vips
COPY --from=builder /usr/local/lib/libvips* /usr/local/lib/
RUN apk update && apk upgrade --available && sync && apk add --no-cache libpng-dev jpeg-dev libwebp-dev tiff-dev zlib-dev
COPY --from=builder /heic2jpeg/heic2jpeg /heic2jpeg
ENTRYPOINT ["/heic2jpeg"]
