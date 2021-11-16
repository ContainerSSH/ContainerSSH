FROM containerssh/agent AS agent

FROM ubuntu:20.04

RUN apt update \
    && apt install -y openssh-server \
    && rm -rf /var/lib/apt/lists/*

RUN mkdir -p /run/sshd

COPY init.sh /init.sh
RUN chmod +x /init.sh

COPY --from=agent /usr/bin/containerssh-agent /usr/bin/containerssh-agent

HEALTHCHECK --start-period=10s CMD /usr/bin/ssh-keyscan 127.0.0.1

EXPOSE 22

CMD ["/init.sh"]