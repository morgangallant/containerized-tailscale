FROM golang:alpine
RUN apk update && apk add bash
RUN go install tailscale.com/cmd/tailscale@v1.6.0
RUN go install tailscale.com/cmd/tailscaled@v1.6.0
ADD . /build/
WORKDIR /build/
RUN go build -o demo .
RUN mkdir -p /tailscale
ARG TAILSCALE_KEY
ENTRYPOINT ["/build/entrypoint.sh"]
