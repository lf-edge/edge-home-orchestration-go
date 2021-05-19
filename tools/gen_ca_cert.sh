#!/bin/bash

mkdir -p ./cert

# CA Root Certificate
# Generate root certificate private key: ca.key
openssl genrsa -out ./cert/ca.key 2048

# Generate a self-signed root certificate: ca.crt
openssl req -new -key ./cert/ca.key -x509 -days 3650 -out ./cert/ca.crt -subj /C=KR/ST=Seoul/O="Samsung Electronics"/CN="Home Edge CA Root"