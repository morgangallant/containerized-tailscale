#!/bin/bash

# The entrypoint for this docker container. This script is responsible for
# configuring Tailscale within this container, and then starting the application.

echo "Starting Demo"

# tailscale{,d} in /home/coder/go/bin

# Run the application.
/home/coder/demo