#!/bin/bash

# The entrypoint for this docker container. This script is responsible for
# configuring Tailscale within this container, and then starting the application.

echo "Starting Demo"

# Starting Tailscaled
(&>/dev/null tailscaled --socks5-server=localhost:1080 \
    --state=/tailscale/tailscale.state \
    --tun=userspace-networking \
    --socket=/tailscale/tailscale.sock &)

echo "Started Tailscaled"

# Authenticate
until tailscale --socket=/tailscale/tailscale.sock \
    up \
    --authkey=$TAILSCALE_KEY
do 
    echo "Waiting for Tailscale Authentication"
    sleep 5
done

echo "Authenticated with Tailscale"

# Run the application.
# Using 'exec' here should forward SIGINT & SIGTERM to program for cleanup.
exec /build/demo
