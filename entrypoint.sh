#!/bin/bash

# The entrypoint for this docker container. This script is responsible for
# configuring Tailscale within this container, and then starting the application.

echo "Starting Demo"

# Starting Tailscaled
/home/coder/go/bin/tailscaled --socks5-server=localhost:1080 \
    --state=/home/coder/.tailscale/tailscale.state \
    --tun=userspace-networking \
    --socket=/home/coder/.tailscale/tailscale.sock

# Run the application.
/home/coder/demo