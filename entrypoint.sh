#!/bin/bash

# The entrypoint for this docker container. This script is responsible for
# configuring Tailscale within this container, and then starting the application.

echo "Starting Demo"

# Starting Tailscaled
/home/coder/go/bin/tailscaled --socks5-server=localhost:1080 \
    --state=/home/coder/.tailscale/tailscale.state \
    --tun=userspace-networking \
    --socket=/home/coder/.tailscale/tailscale.sock &

echo "Started Tailscaled"

# Authenticate
until /home/coder/go/bin/tailscale up \
    --socket=/home/coder/.tailscale/tailscale.sock \
    --authkey=$TAILSCALE_KEY
do 
    echo "Waiting for Tailscale Authentication"
    sleep 1
done

echo "Authenticated with Tailscale"

# Run the application.
/home/coder/demo