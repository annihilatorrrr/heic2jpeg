FROM golang:1.22.5-alpine3.20 as builder
WORKDIR /heic2jpeg
RUN apk update && apk upgrade --available && sync && apk add --no-cache build-base g++ cmake libc-dev libheif-dev libde265-dev
RUN cmake -DCMAKE_CXX_FLAGS="-march=native" .
COPY . .
RUN go build -ldflags="-w -s" .
FROM alpine:3.20.1
# RUN apk update && apk upgrade --available && sync && apk add --no-cache libheif-dev
COPY --from=builder /heic2jpeg/heic2jpeg /heic2jpeg
ENTRYPOINT ["/heic2jpeg"]
