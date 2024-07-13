FROM golang:1.22.5-bookworm as builder
WORKDIR /heic2jpeg
RUN apt-get update && apt-get install -y build-essential g++ cmake libc-dev libheif-dev libde265-dev libjpeg-dev libpng-dev zlib1g-dev libjpeg-dev
COPY . .
RUN go build -ldflags="-w -s" .
FROM debian:latest
# RUN apk update && apk upgrade --available && sync && apk add --no-cache libheif-dev
COPY --from=builder /heic2jpeg/heic2jpeg /heic2jpeg
ENTRYPOINT ["/heic2jpeg"]
