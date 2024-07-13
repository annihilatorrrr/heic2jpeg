# Build stage
FROM golang:1.22.5-alpine3.20 as builder
WORKDIR /heic2jpeg

# Install necessary packages
RUN apk update && apk add --no-cache build-base git cmake make x265-dev g++ libde265-dev jpeg-dev libtool

# Clone and build libheif
RUN git clone https://github.com/strukturag/libheif.git && \
    cd libheif && \
    mkdir build && \
    cd build && \
    cmake .. && \
    make && \
    make install && \
    ldconfig

# Copy source code and build Go project
COPY . .
RUN go build -ldflags="-w -s" .

# Final stage
FROM alpine:3.20.1

# Install only the necessary runtime dependencies
RUN apk update && apk add --no-cache x265 libde265 jpeg

# Copy the built libheif from the builder stage
COPY --from=builder /usr/local/lib/libheif.so* /usr/local/lib/
COPY --from=builder /usr/local/include/libheif /usr/local/include/libheif
COPY --from=builder /usr/local/bin/heif-* /usr/local/bin/

# Copy the built binary from the builder stage
COPY --from=builder /heic2jpeg/heic2jpeg /heic2jpeg

# Update the dynamic linker run-time bindings
RUN ldconfig

# Set entrypoint
ENTRYPOINT ["/heic2jpeg"]
