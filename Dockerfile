FROM golang:alpine
RUN go install tailscale.com/cmd/tailscale@v1.6.0
RUN go install tailscale.com/cmd/tailscaled@v1.6.0
ADD . /build/
WORKDIR /build/
RUN go build -o demo .
ARG TAILSCALE_KEY
RUN mkdir -p /tailscale
ENTRYPOINT ["/build/entrypoint.sh"]
