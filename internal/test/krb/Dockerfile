FROM ubuntu:20.04

RUN echo -e "\e[32mInstalling Kerberos KDC and admin server...\e[0m" && \
    DEBIAN_FRONTEND=noninteractive apt update && \
    DEBIAN_FRONTEND=noninteractive apt install -y krb5-kdc krb5-admin-server && \
    rm -rf /var/lib/apt/lists/*

ADD root /
    
RUN echo -e "\e[32mCreating Kerberos database...\e[0m" && \
    kdb5_util create -r TESTING.CONTAINERSSH.IO -s -P testing

RUN echo -e "\e[32mAdding Kerberos admin user...\e[0m" && \
    kadmin.local -q "addprinc -pw testing admin"

CMD ["/usr/local/bin/init.sh"]

EXPOSE 88
EXPOSE 88/udp
EXPOSE 464
EXPOSE 464/udp
EXPOSE 750
EXPOSE 750/udp
