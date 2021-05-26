#!/bin/bash

if [ $# -ne 1 ]
then
    echo "Generate Home Edge Node (HEN) Certificate"
    echo "Usage:"
    echo "-------------------------------------------------------------------------------"
    echo "  $0 [IP]             : generate hen certificate for node with IP adrress"
    echo "Example:"
    echo "  $0 192.168.0.100    : generate hen certificate for node with 192.168.0.100"
    echo "-------------------------------------------------------------------------------"
    exit 1
fi

mkdir -p ./cert/$1

# Home Edge Node (HEN) Certificate
# Generate HEN Certificate private key : hen.key
openssl genrsa -out ./cert/$1/hen.key 2048

# Generate HEN Certificate request: hen.csr
openssl req -new -nodes -key ./cert/$1/hen.key -out ./cert/$1/hen.csr -subj /C=KR/ST=Seoul/O="Samsung Electronics"/CN="Home Edge Node Certificate"

# Signature HEN Certificate: hen.crt
openssl x509 -req -extfile <(printf "subjectAltName=IP:$1") -days 365 -in ./cert/$1/hen.csr -CA ./cert/ca.crt -CAkey ./cert/ca.key -CAcreateserial -out ./cert/$1/hen.crt