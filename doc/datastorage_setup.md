# ZeroMQ installation on Linux

## Quick start ##
This section provides how to download and install zeromq

#### 1. Install the dependency packages
- sudo apt-get install libtool pkg-config build-essential autoconf   automake

- sudo apt-get install libzmq3-dev

#### 2. Install libsodium (Zero MQ has dependcy on this package)
1. Download the tar from https://libsodium.gitbook.io/doc/installation (Version: 1.0.18)
2. tar -xvf tar -xvf libsodium_1.0.18.orig.tar.gz 
3. cd libsodium-1.0.18
4. ./autogen.sh 
If it succeeds, you can see ./configure file generated
5. ./configure
6. make && make check
7. sudo make install
8. sudo ldconfig

#### 3. Install zeromq
After installing the Dependency packages, zero mq can be isntalled with the following steps

 Download the tar from https://github.com/zeromq/libzmq/tags/v4.2.2

# Unpack tarball package
 tar xvf libzmq-4.2.2.tar.gz
 cd libzmq-4.2.2/
 ./autogen.sh

# Install any other dependency for running configure.sh
 sudo apt-get update
 sudo apt-get install -y libtool pkg-config build-essential autoconf automake uuid-dev

# Create make file
 ./configure
>If this command succeeds you can see a make file generated

# Install the zeromq with make file
8. sudo make install

# Install zeromq driver on linux
9. sudo ldconfig

# Check installed
10. ldconfig -p | grep zmq
> if the installation succeeds then you should see the expected output
# Expected
############################################################
# libzmq.so.5 (libc6,x86-64) => /usr/local/lib/libzmq.so.5
# libzmq.so (libc6,x86-64) => /usr/local/lib/libzmq.so
############################################################
