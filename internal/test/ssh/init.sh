#!/bin/bash

set -e

useradd -rm -d "/home/${SSH_USERNAME}" -s /bin/bash -u 1000 "${SSH_USERNAME}"

echo "${SSH_USERNAME}:${SSH_PASSWORD}" | chpasswd

exec /usr/sbin/sshd -D -o AcceptEnv=*