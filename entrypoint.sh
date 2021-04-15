#!/bin/bash

# The entrypoint for this docker container. This script is responsible for
# configuring Tailscale within this container, and then starting the application.

tailscaled \
--socks5-server=localhost:1080 \
--state=/var/lib/tailscale/tailscale.state \
--tun=userspace-networking \
--socket=/var/lib/tailscale/tailscale.sock

# tailscale --socket=/tailscale/tailscale.sock up --authkey $TAILSCALE_KEY

# # Run the application.
# /mg/demo