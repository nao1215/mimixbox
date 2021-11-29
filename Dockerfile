FROM golang as builder
ENV ROOT=/go/src/app
WORKDIR ${ROOT}

# Install mimixbox
RUN go install github.com/nao1215/mimixbox/cmd/mimixbox@latest
# Create mimixbox symbolic link in container.
RUN mimixbox --full-install /usr/local/bin/

# Set root password
RUN echo 'root:password' | chpasswd

# Create new user
RUN useradd mimixbox -m -s /bin/bash
RUN echo 'mimixbox:password' |chpasswd

# If you want to administrator privileges, you become the root user.
# RUN echo "mimixbox    ALL=(ALL)       ALL" >> /etc/sudoers
CMD ["su", "-", "mimixbox"]