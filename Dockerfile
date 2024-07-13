FROM golang:1.22.5-slim-bookworm
WORKDIR /heic2jpeg
RUN apt update && apt upgrade -y
COPY . .
RUN go build -ldflags="-w -s" .
FROM debian:bookworm-slim
RUN apt update && apt upgrade -y
RUN apt install -y ffmpeg apt-utils build-essential git cmake make pkg-config libx265-dev libde265-dev libjpeg-dev libtool libheif-dev
RUN git clone https://github.com/strukturag/libheif.git && \
    cd libheif && \
    mkdir build && cd build && \
    cmake .. && \
    make && \
    make install
ENV LD_LIBRARY_PATH=/usr/local/lib:$LD_LIBRARY_PATH
COPY --from=builder /heic2jpeg/heic2jpeg /heic2jpeg
ENTRYPOINT ["/heic2jpeg"]
