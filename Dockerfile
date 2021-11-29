FROM golang as builder
ENV ROOT=/go/app
WORKDIR ${ROOT}

# 1) Install package for setup
# 2) Install mimixbox
# 3) Create mimixbox symbolic link in container.
# 4) Setting root user password
# 5) Add mimixbox user
# 6) Setting mimixbox user password
RUN apt-get update && apt-get install -y expect  && \   
    go install github.com/nao1215/mimixbox/cmd/mimixbox@latest  && \
    mimixbox --full-install /usr/local/bin/
RUN echo 'root:password' | chpasswd
RUN useradd mimixbox -m -s /bin/bash &&\
    echo 'mimixbox:password' |chpasswd

# Copy ShellSpec installer
COPY ./scripts/installShellSpec.sh /home/mimixbox
RUN  /home/mimixbox/installShellSpec.sh && rm /home/mimixbox/install*.sh

# If you want to administrator privileges, you become the root user.
# RUN echo "mimixbox    ALL=(ALL)       ALL" >> /etc/sudoers
CMD ["su", "-", "mimixbox"]