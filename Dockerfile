FROM golang:1.22.5-alpine3.20 as builder
WORKDIR /heic2jpeg
RUN apk update && apk upgrade --available && sync && apk add --no-cache build-base git cmake make x265-dev libde265-dev jpeg-dev libtool
RUN git clone https://github.com/strukturag/libheif.git && cd libheif && mkdir build && cd build && cmake .. && make && make install && ldconfig
COPY . .
RUN go build -ldflags="-w -s" .
FROM alpine:3.20.1
RUN apk update && apk upgrade --available && sync && apk add --no-cache build-base git cmake make x265-dev libde265-dev jpeg-dev libtool
RUN git clone https://github.com/strukturag/libheif.git && cd libheif && mkdir build && cd build && cmake .. && make && make install && ldconfig
RUN apk del git cmake make build-base
COPY --from=builder /heic2jpeg/heic2jpeg /heic2jpeg
ENTRYPOINT ["/heic2jpeg"]
