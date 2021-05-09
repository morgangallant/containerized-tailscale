# Build Step
FROM golang:alpine AS build
RUN go install tailscale.com/cmd/tailscale@v1.6.0
RUN go install tailscale.com/cmd/tailscaled@v1.6.0
ADD . /build/
WORKDIR /build/
RUN go build -o demo .

# Run Step
FROM golang:alpine
RUN mkdir -p /run/tailscale/
WORKDIR /run/
COPY --from=build /build/demo /run/demo
COPY --from=build /build/entrypoint.sh /run/entrypoint.sh
ARG TAILSCALE_KEY
ENTRYPOINT ["/run/entrypoint.sh"]

