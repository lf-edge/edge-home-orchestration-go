#!/bin/bash

echo The platform is $TARGETPLATFORM.
if [ $TARGETPLATFORM = "linux/amd64" ]; then
    arch="amd64"
elif [ $TARGETPLATFORM = "linux/386" ]; then
    arch="386"
elif [ $TARGETPLATFORM = "linux/arm64" ]; then
    arch="arm64"
elif [ $TARGETPLATFORM = "linux/arm/v7" ]; then
    arch="armv6l"
fi

wget --no-check-certificate https://golang.org/dl/go$GOVERSION.linux-$arch.tar.gz && \
tar -C /usr/local -xzf go$GOVERSION.linux-$arch.tar.gz && \
ln -s $GOPATH/bin/go /usr/bin/
