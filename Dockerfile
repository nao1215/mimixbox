FROM golang as builder
ENV ROOT=/go/app
ENV IT_SHELL=/home/mimixbox/do_integration_test.sh
WORKDIR ${ROOT}

# 1) Setting root user password
# 2) Add mimixbox user
# 3) Setting mimixbox user password
RUN echo 'root:password' | chpasswd
RUN useradd mimixbox -m -s /bin/bash &&\
    echo 'mimixbox:password' |chpasswd
RUN apt-get update && apt-get upgrade && apt-get -y install sudo file

# Copy ShellSpec installer
COPY ./test/it /home/mimixbox/integration_tests
RUN  git clone https://github.com/shellspec/shellspec.git && \
    cd shellspec && make install

RUN echo "#!/bin/bash" > ${IT_SHELL} && \
    echo "cd /home/mimixbox/integration_tests && shellspec\n" >> ${IT_SHELL} && \
    chmod a+x ${IT_SHELL} && \
    chown -R mimixbox:mimixbox /home/mimixbox/.

RUN git clone https://github.com/nao1215/mimixbox.git && cd mimixbox && \
    make build
RUN cd ${ROOT}/mimixbox && sudo make full-install

# If you want to administrator privileges, you become the root user.
# RUN echo "mimixbox    ALL=(ALL)       ALL" >> /etc/sudoers
CMD ["su", "-", "mimixbox"]