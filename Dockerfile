FROM golang:1.16

# Build the Application
ADD . /mg 
WORKDIR /mg
RUN go build -o demo .

# Install Tailscale
RUN curl -fsSL https://pkgs.tailscale.com/stable/ubuntu/bionic.gpg | apt-key add -
RUN curl -fsSL https://pkgs.tailscale.com/stable/ubuntu/bionic.list | tee /etc/apt/sources.list.d/tailscale.list
RUN apt-get update
RUN apt-get install tailscale -y

ARG TAILSCALE_KEY

# Configure Entrypoint
# ENTRYPOINT ["/mg/entrypoint.sh"]
