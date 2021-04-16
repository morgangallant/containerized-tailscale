FROM golang:1.16

# Configure user.
RUN adduser --gecos '' --disabled-password coder && \
    mkdir -p /etc/sudoers.d && \
    echo "coder ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers.d/nopasswd
RUN ARCH="$(dpkg --print-architecture)" && \
    curl -fsSL "https://github.com/boxboat/fixuid/releases/download/v0.5/fixuid-0.5-linux-$ARCH.tar.gz" | tar -C /usr/local/bin -xzf - && \
    chown root:root /usr/local/bin/fixuid && \
    chmod 4755 /usr/local/bin/fixuid && \
    mkdir -p /etc/fixuid && \
    printf "user: coder\ngroup: coder\n" > /etc/fixuid/config.yml
WORKDIR /home/coder
USER coder

# Build the Application
ADD . /home/coder/
RUN go build -o demo .

# Install Tailscale
RUN curl -fsSL https://pkgs.tailscale.com/stable/ubuntu/bionic.gpg | sudo apt-key add -
RUN curl -fsSL https://pkgs.tailscale.com/stable/ubuntu/bionic.list | sudo tee /etc/apt/sources.list.d/tailscale.list
RUN sudo apt-get update
RUN sudo apt-get install tailscale -y

# Get the Tailscale Key from the Environment.
ARG TAILSCALE_KEY

# Make a folder for the tailcale sock/state.
RUN mkdir -p .tailscale/

# Configure Entrypoint
ENTRYPOINT ["/mg/entrypoint.sh"]
