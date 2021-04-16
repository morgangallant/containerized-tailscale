FROM debian:10

# Configure basics.
RUN apt-get update \
 && apt-get install -y \
    curl \
    wget \
    git \
    sudo \
  && rm -rf /var/lib/apt/lists/*

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

# Install Go
RUN wget https://golang.org/dl/go1.16.3.linux-amd64.tar.gz
RUN sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.16.3.linux-amd64.tar.gz
RUN rm go1.16.3.linux-amd64.tar.gz

# Build the Application
ADD . /home/coder/
RUN /usr/local/go/bin/go build -o demo .

# Install Tailscale
RUN /usr/local/go/bin/go install tailscale.com/cmd/tailscale@v1.6.0
RUN /usr/local/go/bin/go install tailscale.com/cmd/tailscaled@v1.6.0

# Get the Tailscale Key from the Environment.
ARG TAILSCALE_KEY

# Make a folder for the tailcale sock/state.
RUN mkdir -p .tailscale/

# Configure Entrypoint
ENTRYPOINT ["/home/coder/entrypoint.sh"]
