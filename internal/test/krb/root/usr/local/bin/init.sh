#!/bin/bash

set -m
set -e
set -o monitor

if [ -z "${KERBEROS_USERNAME}" ]; then
    echo -e "\e[33mThe KERBEROS_USERNAME environment variable is not set.\e[0m"
    exit 1
fi

if [ -z "${KERBEROS_PASSWORD}" ]; then
    echo -e "\e[33mThe KERBEROS_PASSWORD environment variable is not set.\e[0m"
    exit 1
fi

rm -rf /var/lib/krb5kdc
mkdir -p /var/lib/krb5kdc

echo -e "\e[32mCreating Kerberos database...\e[0m"
kdb5_util create -r TESTING.CONTAINERSSH.IO -s -P testing

echo -e "\e[32mAdding Kerberos admin user...\e[0m"
kadmin.local -q "addprinc -pw ${KERBEROS_PASSWORD} ${KERBEROS_USERNAME}"
echo -n "" >/etc/krb5kdc/kadm5.acl
echo "${KERBEROS_USERNAME}@TESTING.CONTAINERSSH.IO *" >>/etc/krb5kdc/kadm5.acl

trap finish SIGCHLD EXIT
finish() {
    echo -e "\e[33mSignal received, exiting...\e[0m"
    exit 1
}

echo -e "\e[32mStarting kadmind...\e[0m"
kadmind -nofork &

echo -e "\e[32mStarting KDC...\e[0m"
krb5kdc -n &

echo -e "\e[32mWaiting for processes to stop...\e[0m"
wait -n

echo -e "\e[32mOne or more processes stopped, quitting...\e[0m"
exit $?
