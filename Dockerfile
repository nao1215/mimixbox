# Pin the base image to a specific tag so builds are reproducible and not
# affected by a silently moving `golang:latest`. Matches the toolchain declared
# in go.mod (go 1.25.x).
FROM golang:1.25-bookworm AS builder
ENV ROOT=/go/app
ENV IT_SHELL=/home/mimixbox/do_integration_test.sh
# Pin atago to a tagged release for reproducible integration tests.
ENV ATAGO_VERSION=v0.17.0
WORKDIR ${ROOT}

# 1) Setting root user password
# 2) Add mimixbox user
# 3) Setting mimixbox user password
RUN echo 'root:password' | chpasswd
RUN useradd mimixbox -m -s /bin/bash &&\
    echo 'mimixbox:password' |chpasswd
RUN apt-get update && \
    apt-get -y install --no-install-recommends sudo file && \
    rm -rf /var/lib/apt/lists/*

# Install atago (pinned tag) for the integration tests. atago requires a newer
# Go than this image's toolchain, so GOTOOLCHAIN=auto lets `go install` fetch
# the toolchain atago declares while mimixbox itself keeps building with the
# image's pinned one.
RUN env GOTOOLCHAIN=auto go install "github.com/nao1215/atago@${ATAGO_VERSION}" && \
    install -m 0755 /go/bin/atago /usr/local/bin/atago

# Build MimixBox from the local source tree (not a remote clone) so the image
# always reflects the working copy, with cgo enabled in the toolchain image.
COPY . ${ROOT}/mimixbox
RUN cd ${ROOT}/mimixbox && make build && sudo make full-install

# Make the integration tests available to the mimixbox user. The applets are
# already full-installed into /usr/local/bin, so atago runs the specs against
# them directly (sequential, like `make e2e` — the kill-family scenarios
# signal processes by name and must not race other scenarios).
COPY ./e2e/atago /home/mimixbox/integration_tests
RUN echo "#!/bin/bash" > ${IT_SHELL} && \
    echo "atago run --parallel 1 /home/mimixbox/integration_tests" >> ${IT_SHELL} && \
    chmod a+x ${IT_SHELL} && \
    chown -R mimixbox:mimixbox /home/mimixbox/.

# If you want administrator privileges, become the root user.
# RUN echo "mimixbox    ALL=(ALL)       ALL" >> /etc/sudoers
CMD ["su", "-", "mimixbox"]
