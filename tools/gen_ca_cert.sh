#!/bin/bash

mkdir -p ./certs

# CA Root Certificate
# Generate root private key: ca-key.pem
openssl genrsa -out ./certs/ca-key.pem 2048

# Generate a self-signed root certificate: ca-crt.pem
openssl req -new -key ./certs/ca-key.pem -x509 -days 3650 -out ./certs/ca-crt.pem -outform PEM -subj /C=KR/ST=Seoul/O="Samsung Electronics"/CN="Home Edge CA Root"
