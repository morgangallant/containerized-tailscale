FROM golang:1.16

# Build the Application
ADD . /mg 
WORKDIR /mg
RUN go build -o demo .

# Install Tailscale
RUN curl -fsSL https://pkgs.tailscale.com/stable/ubuntu/bionic.gpg | sudo apt-key add -
RUN curl -fsSL https://pkgs.tailscale.com/stable/ubuntu/bionic.list | sudo tee /etc/apt/sources.list.d/tailscale.list
RUN sudo apt-get update
RUN sudo apt-get install tailscale

# Configure Entrypoint
# ENTRYPOINT ["/mg/entrypoint.sh"]
