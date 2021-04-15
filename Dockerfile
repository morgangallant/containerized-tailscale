FROM golang:1.16

# Build the Application
ADD . /mg 
WORKDIR /mg
RUN go build -o demo .

# Install Tailscale
RUN go install tailscale.com/cmd/tailscale@v1.6.0
RUN go install tailscale.com/cmd/tailscaled@v1.6.0
RUN mkdir /tailscale # Will be used to store Tailscale data.

# Configure Entrypoint
ENTRYPOINT ["/mg/entrypoint.sh"]
