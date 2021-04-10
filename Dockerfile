ARG GO_VERSION=1.16
ARG ALPINE_VERSION=3.13

ARG GOOS=linux
ARG GOARCH=amd64

FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:${GO_VERSION}-alpine${ALPINE_VERSION} as build

ARG GOOS
ARG GOARCH

# Provide CC and CXX environment variables with the C and C++ compiler to use
ARG CC
ARG CXX

# Install libc, musl and C compilers
RUN apk --no-cache add gcc g++ make libc-dev clang mingw-w64-gcc

# Use the project root as the current workdir
WORKDIR /go/src/github.com/peerbridge/peerbridge

# Copy go.mod into the container's workspace
COPY go.mod .

# Install dependencies
RUN go mod download

# Copy the local package files to the container's workspace
COPY . .

# Build the peerbridge app inside the container
RUN \
    BIN="/go/bin/peerbridge-${GOOS}-${GOARCH}"; \
    if [ "${GOOS}" == "windows" ]; then \
        BIN="${BIN}.exe"; \
    fi; \
    CC=${CC} CXX=${CXX} GOOS=${GOOS} GOARCH=${GOARCH} CGO_ENABLED=1 go install std; \
    CC=${CC} CXX=${CXX} GOOS=${GOOS} GOARCH=${GOARCH} CGO_ENABLED=1 go build -a -o ${BIN} -ldflags="-extldflags '-static' -s -w"

FROM scratch as bin-unix

ARG GOOS
ARG GOARCH

# Copy the peerbridge executable into the image
COPY --from=build /go/bin/peerbridge-${GOOS}-${GOARCH} /

FROM bin-unix as bin-linux

FROM bin-unix as bin-darwin

FROM scratch as bin-windows

ARG GOOS
ARG GOARCH

# Copy the peerbridge executable into the image
COPY --from=build /go/bin/peerbridge-${GOOS}-${GOARCH}.exe /

FROM bin-${GOOS} AS bin

# Copy the peerbridge executable into the image
COPY --from=build /go/bin/${BIN} /

FROM alpine:${ALPINE_VERSION}

LABEL org.opencontainers.image.source https://github.com/peerbridge/peerbridge

ARG GOOS
ARG GOARCH

# Copy the peerbridge executable into a alpine image.
COPY --from=build /go/bin/peerbridge-${GOOS}-${GOARCH} /usr/local/bin/peerbridge

# Copy the template files into the alpine image.
COPY --from=build /go/src/github.com/peerbridge/peerbridge/templates ./templates

# Copy the static files into the alpine image.
COPY --from=build /go/src/github.com/peerbridge/peerbridge/static ./static

# Start peerbridge app inside the container.
ENTRYPOINT [ "peerbridge", "server" ]

# Expose default server port.
EXPOSE 8080

# Expose default p2p port.
EXPOSE 9080
