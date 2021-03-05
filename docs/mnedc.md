# MNEDC (Multi NAT Edge Discovery and Communication)
## Contents
1. [Introduction](#1-introduction)
2. [MNEDC Server](#2-mnedc-server)
3. [MNEDC Client](#3-mnedc-client)
4. [How to Setup](#4-how-to-setup)  
    4.1 [Setting up the MNEDC Server](#41-setting-up-the-mnedc-server)      
    4.2 [Setting up the MNEDC Client](#42-setting-up-the-mnedc-client) 

## 1. Introduction
When devices are connected under multiple levels of NATs (or routers), the devices under upper level of NAT cannot reach the devices under the lower ones directly even if they know the IP address (Network limitation). 
With the MNEDC system, the devices can communicate with each other once they establish a persistent connection with MNEDC server which acts as a relay between devices and helps in communication by providing a channel.

## 2. MNEDC Server
This is a TCP server running on any one of the devices which is reachable from all other IoT devices in the network. Since Sub NAT devices can reach Main NAT device and the other way round is not possible, we need to
run the MNEDC server in the Main NAT. This MNEDC Server registers the client whenever the request comes and establishes a persistent connection. Also, it provides it with a unique virtual IP in the address space of 
10.0.0.1~255. Then it maintains a key-value map of the unique virtual IP and the TCP connection object. Now the role of server is to get the packets from clients, extract the target virtual IP, and write the packets 
on the TCP connection object which is retrieved from the map. In this way a device anywhere in the network can communicate with all the devices irrespective of their position in network.

## 3. MNEDC Client
The job of the client is to first create a TCP connection with the MNEDC Server and upon receipt of the virtual IP, create a tun interface and assign the IP as given by the MNEDC Server. Then the client reads all the
packets on the tun interface and write those packets on the TCP connection established with the server, and capture the packets on the TCP connection and write those on the TUN interface. In this way applications using
virtual IP to communicate with the peers will be able to send and receive the packets.

## 4. How to Setup
### 4.1 Setting up the MNEDC Server
Just run the following command to run the MNEDC server on the device:
`./build.sh mnedcserver`.
Note that there should be only one device running the MNEDC Server in the network.

### 4.2 Setting up the MNEDC Client
Steps to run the MNEDC Client:
1. Edit the client-config.yaml file inside /configs/mnedc/ directory and put the IP address of the device which is running the MNEDC server.
2. Copy this client-config.yaml file to /var/edge-orchestration/mnedc folder.
3. Run the following command:
`./build.sh mnedcclient`


