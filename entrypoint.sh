#!/bin/bash

# The entrypoint for this docker container. This script is responsible for
# configuring Tailscale within this container, and then starting the application.

echo "Starting Demo"

# Run the application.
/home/coder/mg/demo