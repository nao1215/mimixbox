FROM golang as builder
ENV ROOT=/go/app
ENV IT_SHELL=/home/mimixbox/do_integration_test.sh
WORKDIR ${ROOT}

# 1) Install mimixbox
# 2) Create mimixbox symbolic link in container.
# 3) Setting root user password
# 4) Add mimixbox user
# 5) Setting mimixbox user password
RUN go install github.com/nao1215/mimixbox/cmd/mimixbox@latest  && \
    mimixbox --full-install /usr/local/bin/
RUN echo 'root:password' | chpasswd
RUN useradd mimixbox -m -s /bin/bash &&\
    echo 'mimixbox:password' |chpasswd

# Copy ShellSpec installer
COPY ./test /home/mimixbox/integration_tests
RUN  git clone https://github.com/shellspec/shellspec.git && \
    cd shellspec && make install

RUN echo "#!/bin/bash" > ${IT_SHELL} && \
    echo "cd /home/mimixbox/integration_tests && shellspec\n" >> ${IT_SHELL} && \
    chmod a+x ${IT_SHELL}

# If you want to administrator privileges, you become the root user.
# RUN echo "mimixbox    ALL=(ALL)       ALL" >> /etc/sudoers
CMD ["su", "-", "mimixbox"]