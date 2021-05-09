FROM golang:alpine

# Tailscaled can't run as root user, cuz of an SO_MARK issue.
# See: https://github.com/tailscale/tailscale/issues/634.
ARG USER=default
ENV HOME /home/$USER
RUN apk add --update sudo bash
RUN adduser -D $USER \
  && echo "$USER ALL=(ALL) NOPASSWD: ALL" > /etc/sudoers.d/$USER \
  && chmod 0440 /etc/sudoers.d/$USER
USER $USER
WORKDIR $HOME

# Install Tailscale & Tailscaled.
ARG TAILSCALE_KEY
RUN mkdir -p $HOME/tailscale
RUN go install tailscale.com/cmd/tailscale@v1.6.0
RUN go install tailscale.com/cmd/tailscaled@v1.6.0

# Build the application.
ADD . $HOME/build
WORKDIR $HOME/build
RUN go build -o demo .

# Fire up Tailscaled, Authenticate, and Run.
ENTRYPOINT ["$HOME/build/entrypoint.sh"]
