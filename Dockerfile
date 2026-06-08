FROM golang AS builder
ENV ROOT=/go/app
ENV IT_SHELL=/home/mimixbox/do_integration_test.sh
WORKDIR ${ROOT}

# 1) Setting root user password
# 2) Add mimixbox user
# 3) Setting mimixbox user password
RUN echo 'root:password' | chpasswd
RUN useradd mimixbox -m -s /bin/bash &&\
    echo 'mimixbox:password' |chpasswd
RUN apt-get update && apt-get upgrade -y && \
    apt-get -y install sudo file libpam0g-dev

# Install ShellSpec for the integration tests.
RUN git clone https://github.com/shellspec/shellspec.git && \
    cd shellspec && make install

# Build MimixBox from the local source tree (not a remote clone) so the image
# always reflects the working copy, with cgo enabled in the toolchain image.
COPY . ${ROOT}/mimixbox
RUN cd ${ROOT}/mimixbox && make build && sudo make full-install

# Make the integration tests available to the mimixbox user.
COPY ./test/it /home/mimixbox/integration_tests
RUN echo "#!/bin/bash" > ${IT_SHELL} && \
    echo "cd /home/mimixbox/integration_tests && shellspec" >> ${IT_SHELL} && \
    chmod a+x ${IT_SHELL} && \
    chown -R mimixbox:mimixbox /home/mimixbox/.

# If you want administrator privileges, become the root user.
# RUN echo "mimixbox    ALL=(ALL)       ALL" >> /etc/sudoers
CMD ["su", "-", "mimixbox"]
