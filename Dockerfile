FROM golang:1.22.5-bookworm as builder
WORKDIR /heic2jpeg
RUN apt update && apt upgrade -y
COPY . .
RUN go build -ldflags="-w -s" .
FROM debian:latest
RUN apt update && apt upgrade -y
RUN apt install ffmpeg apt-utils build-essential pkg-config libx265-dev libde265-dev libjpeg-dev libtool libheif-dev -y
COPY --from=builder /heic2jpeg/heic2jpeg /heic2jpeg
ENTRYPOINT ["/heic2jpeg"]
