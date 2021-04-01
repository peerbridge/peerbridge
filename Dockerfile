FROM golang:1.16.2-alpine3.13 as build
# Copy the local package files to the container's workspace.
ADD . /go/src/github.com/peerbridge/peerbridge
# Use the project root as the current workdir
WORKDIR /go/src/github.com/peerbridge/peerbridge
# Install git
RUN apk --no-cache add git
# Install build tools together with gcc
RUN apk --no-cache add build-base
# Get dependencies
RUN go get -v -t -d ./...
# Build the peerbridge app inside the container.
RUN go install

FROM alpine:3.13
# Copy the peerbridge executable into the alpine image.
COPY --from=build /go/bin/peerbridge /usr/local/bin/peerbridge
# Copy the template files into the alpine image.
COPY --from=build /go/src/github.com/peerbridge/peerbridge/templates ./templates
# Copy the static files into the alpine image.
COPY --from=build /go/src/github.com/peerbridge/peerbridge/static ./static
# Start peerbridge app inside the container.
ENTRYPOINT [ "peerbridge" ]
# Expose default server port.
EXPOSE 8080
