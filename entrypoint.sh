#!/bin/bash

# The entrypoint for this docker container. This script is responsible for
# configuring Tailscale within this container, and then starting the application.

# See https://github.com/tailscale/tailscale/issues/634.
rm -rf /tmp/tailscaled
mkdir -p /tmp/tailscaled
chown irc.irc /tmp/tailscaled
rm -rf /var/run/tailscale
mkdir -p /var/run/tailscale
chown irc.irc /var/run/tailscale
cp /var/lib/tailscaled/tailscaled.state /tmp/tailscaled/tailscaled.state
chown irc.irc /tmp/tailscaled/tailscaled.state

nohup sudo -u irc tailscaled \
--tun=userspace-networking \
--socks5-server=localhost:1080 \
--state=/tmp/tailscaled/tailscaled.state \
--socket=/var/run/tailscale/tailscaled.sock --port 41641 &

until tailscale up --authkey $TAILSCALE_KEY; do sleep 1; done

# # Run the application.
/mg/demo