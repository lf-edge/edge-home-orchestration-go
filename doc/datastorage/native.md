# ZeroMQ installation on Linux

## Quick start ##
This section provides how to download and install zeromq

#### 1. Install the dependency packages
```
sudo apt-get install libtool pkg-config build-essential autoconf automake
sudo apt-get install libzmq3-dev
```

#### 2. Install libsodium (Zero MQ has dependcy on this package)

Download the tar from https://libsodium.gitbook.io/doc/installation (Version: 1.0.18)
```
tar -xvf tar -xvf libsodium_1.0.18.orig.tar.gz 
cd libsodium-1.0.18
./autogen.sh 
```

If it succeeds, you can see ./configure file generated
```
./configure
make && make check
sudo make install
sudo ldconfig
```

#### 3. Install zeromq
After installing the Dependency packages, zero mq can be isntalled with the following steps

Download the tar from https://github.com/zeromq/libzmq/tags/v4.2.2

# Unpack tarball package
```
tar xvf libzmq-4.2.2.tar.gz
cd libzmq-4.2.2/
./autogen.sh
```

# Install any other dependency for running configure.sh
```
sudo apt-get update
sudo apt-get install -y libtool pkg-config build-essential autoconf automake uuid-dev
```

# Create make file
```
./configure
```

>If this command succeeds you can see a make file generated

# Install the zeromq with make file
```
sudo make install
```

# Install zeromq driver on linux
```
sudo ldconfig
```

# Check installed
```
ldconfig -p | grep zmq
```

> if the installation succeeds then you should see the expected output

# Expected
```
libzmq.so.5 (libc6,x86-64) => /usr/local/lib/libzmq.so.5
libzmq.so (libc6,x86-64) => /usr/local/lib/libzmq.so
```
