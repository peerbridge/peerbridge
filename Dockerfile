FROM golang:1.14-alpine3.12 as build

# Copy the local package files to the container's workspace.
ADD . /go/src/github.com/peerbridge/peerbridge

# Install git
RUN apk --no-cache add git

# Get dependencies
RUN go get -v -t -d github.com/peerbridge/peerbridge/...

# Build the peerbridge app inside the container.
RUN GCGO_ENABLED=0 go install github.com/peerbridge/peerbridge

FROM alpine:3.12

# Copy the peerbridge executable into a alpine image.
COPY --from=build /go/bin/peerbridge /usr/local/bin/peerbridge

# Start peerbridge app inside the container.
ENTRYPOINT [ "peerbridge" ]

# Expose default server port.
EXPOSE 8000