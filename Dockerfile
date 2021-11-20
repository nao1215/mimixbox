FROM golang

# Create mimixbox symbolic link in container.
COPY mimixbox /usr/local/bin/
RUN mimixbox --full-install /usr/local/bin/

# Set root password
RUN echo 'root:password' | chpasswd

# Create new user
RUN useradd mimixbox -m 
RUN echo 'mimixbox:password' |chpasswd

# If you want to administrator privileges, you become the root user
# RUN echo "mimixbox    ALL=(ALL)       ALL" >> /etc/sudoers

CMD ["su", "-", "mimixbox"]