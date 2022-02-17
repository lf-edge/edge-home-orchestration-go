# Cloud Synchronization
## Contents
1. [Introduction](#1-Introduction)
2. [Launching and Configuring EC2 Instance](#2-Launching-and-Configuring-EC2-Instance)
3. [Installing MQTT Broker on AWS](#3-Installing-MQTT-Broker-on-AWS)
4. [Certificate Generation for MQTT](#4-Certificate-Generation-for-MQTT)
5. [Configurations for MQTT broker](#5-Configurations-for-MQTT)
6. [Run CloudSync](#6-Run-Cloudsync)

# 1. Introduction
This module would be responsible for sending the data (can be sensor data or any reading or image/video data, etc.) collected from different devices in the home environment to a cloud Endpoint (AWS). This Sync to the cloud will be using a API Call by the service application. The broker used is mosquitto which is setup at AWS endpoint. The Service Application will make a POST API call which will then check first for the connection to the broker and then establish a connection if not present to further publish the data to the cloud. After successful publish or in case of failure corresponding response is sent back to the service application<br><br>
<img src="images/CloudSync/MQTTArchitecture.png" alt="image" align="left"/>

### MQTT Example
Consider a Home Scenario where a thermostat sends its temperature update to Broker on topic home/temperature. Two other devices have subscribed to same topic say home/temperature. So The subscribed clients can get update everytime the thermostat publishes data on the topic

![Role of MQTT](images/CloudSync/RoleOfMQTT.png)

# 2. Launching and Configuring EC2 Instance

**This chapter is an example of setting up an AWS EC2 instance for ease of understanding. See Official AWS instructions.**

### A. Create a Role for IOT Access
1. Go to the AWS Web Console and click create a Role
<img src="images/CloudSync/EC2-1.png" alt="image" style="margin:10px"/> 
2. Select EC2 and click on Next: Permissions 
<img src="images/CloudSync/EC2-2.png" alt="image" style="margin:10px"/>
3. Filter with the value AWSIoTConfigAccess. Then select the policy AWSIoTConfigAccess and click on Next: Tags. Skip the next screen by clicking on Next: Review.
<img src="images/CloudSync/EC2-3.png" alt="image" style="margin:10px"/>
4. Enter AWS_IoT_Config_Access as the Role name and enter a Role description. Review the role and click on Create role
<img src="images/CloudSync/EC2-4.png" alt="image" style="margin:10px"/>
5. Now that the Role has been created you can go to Amazon EC2.

### B. Creating EC2 Instance with role created

1. Choose a region, in this article I am using N. Virginia (us-east-1). Then click on Launch Instance and use the filter with the value “ubuntu”. Select the Ubuntu Server 18.04 LTS x86 
<img src="images/CloudSync/EC2-5.png" alt="image" style="margin:10px"/>
2. Select the t2.micro instance type 
<img src="images/CloudSync/EC2-6.png" alt="image" style="margin:10px"/>
3. Click on Next: Configure Instance Details. In the IAM Role dropdown, select AWS_IoT_Config_Access
Make sure you use the default VPC and that the Auto-assign Public IP is Enable to get a public IP automatically. If you wish to use another VPC, make sure the subnet you choose will enable you to remotely connect to your Amazon EC2 instance. Then, click on Next: Add Storage.
<img src="images/CloudSync/EC2-7.png" alt="image" style="margin:10px"/>
4. Leave everything as is and click on Next: Tag Instance. You may assign a tag to your instance. Click on Next: Configure Security Groups. Create a new security group as described in the screenshot
<img src="images/CloudSync/EC2-8.png" alt="image" style="margin:10px"/>
5. Review and launch the EC2 instance. Make sure to select an existing Key Pair or to create a new one in order to connect to the Amazon EC2 instance later on. Once the Amazon EC2 instance is running, click on “Connect” and follow instructions to establish a connection through a terminal.
<img src="images/CloudSync/EC2-9.png" alt="image" style="margin:10px"/>


# 3. Installing MQTT Broker on AWS

Once logged into the Amazon EC2 instance type the following commands:
#Update the list of repositories with one containing the latest version of #Mosquitto and update the package lists
```
sudo apt-add-repository ppa:mosquitto-dev/mosquitto-ppa
sudo apt-get update

```
#Install the Mosquitto broker, Mosquitto clients 
```
sudo apt-get install mosquitto
sudo apt-get install mosquitto-clients
```

# 4. Certificate Generation for MQTT

### Create a basic configuration file:
```
$ touch openssl-ca.cnf
```
Then, add the following to it:
```
HOME            = .
RANDFILE        = $ENV::HOME/.rnd

####################################################################
[ ca ]
default_ca    = CA_default      # The default ca section

[ CA_default ]

default_days     = 1000         # How long to certify for
default_crl_days = 30           # How long before next CRL
default_md       = sha256       # Use public key default MD
preserve         = no           # Keep passed DN ordering

x509_extensions = ca_extensions # The extensions to add to the cert

email_in_dn     = no            # Don't concat the email in the DN
copy_extensions = copy          # Required to copy SANs from CSR to cert

####################################################################
[ req ]
default_bits       = 4096
default_keyfile    = cakey.pem
distinguished_name = ca_distinguished_name
x509_extensions    = ca_extensions
string_mask        = utf8only

####################################################################
[ ca_distinguished_name ]
countryName         = Country Name (2 letter code)
countryName_default = US

stateOrProvinceName         = State or Province Name (full name)
stateOrProvinceName_default = Maryland

localityName                = Locality Name (eg, city)
localityName_default        = Baltimore

organizationName            = Organization Name (eg, company)
organizationName_default    = Test CA, Limited

organizationalUnitName         = Organizational Unit (eg, division)
organizationalUnitName_default = Server Research Department

commonName         = Common Name (e.g. server FQDN or YOUR name) /** In this case you can use *.compute-1.amazonaws.com ***/
commonName_default = Test CA

emailAddress         = Email Address
emailAddress_default = test@example.com

####################################################################
[ ca_extensions ]

subjectKeyIdentifier   = hash
authorityKeyIdentifier = keyid:always, issuer
basicConstraints       = critical, CA:true
keyUsage               = keyCertSign, cRLSign
```
Then, execute the following. The -nodes omits the password or passphrase so you can examine the certificate.
```
$ openssl req -x509 -config openssl-ca.cnf -newkey rsa:4096 -sha256 -nodes -out cacert.pem -outform PEM
```
After the command executes, cacert.pem will be your certificate for CA operations, and cakey.pem will be the private key. Recall the private key does not have a password or passphrase

### Create Another Configuration file using
```
$ touch openssl-server.cnf
```
Then open it, and add the following.
```
HOME            = .
RANDFILE        = $ENV::HOME/.rnd

####################################################################
[ req ]
default_bits       = 2048
default_keyfile    = serverkey.pem
distinguished_name = server_distinguished_name
req_extensions     = server_req_extensions
string_mask        = utf8only

####################################################################
[ server_distinguished_name ]
countryName         = Country Name (2 letter code)
countryName_default = US

stateOrProvinceName         = State or Province Name (full name)
stateOrProvinceName_default = MD

localityName         = Locality Name (eg, city)
localityName_default = Baltimore

organizationName            = Organization Name (eg, company)
organizationName_default    = Test Server, Limited

commonName           = Common Name (e.g. server FQDN or YOUR name)
commonName_default   = Test Server

emailAddress         = Email Address
emailAddress_default = test@example.com

####################################################################
[ server_req_extensions ]

subjectKeyIdentifier = hash
basicConstraints     = CA:FALSE
keyUsage             = digitalSignature, keyEncipherment
subjectAltName       = @alternate_names
nsComment            = "OpenSSL Generated Certificate"

####################################################################
[ alternate_names ]

DNS.1  = example.com
DNS.2  = www.example.com
DNS.3  = mail.example.com
DNS.4  = ftp.example.com

```
Then, create the server certificate request
```
$ openssl req -config openssl-server.cnf -newkey rsa:2048 -sha256 -nodes -out servercert.csr -outform PEM
```
After this command executes, you will have a request in servercert.csr and a private key in serverkey.pem

Next, you have to sign it with your CA.open openssl-ca.cnf and add the following two sections.
```
####################################################################
[ signing_policy ]
countryName            = optional
stateOrProvinceName    = optional
localityName           = optional
organizationName       = optional
organizationalUnitName = optional
commonName             = supplied
emailAddress           = optional

####################################################################
[ signing_req ]
subjectKeyIdentifier   = hash
authorityKeyIdentifier = keyid,issuer
basicConstraints       = CA:FALSE
keyUsage               = digitalSignature, keyEncipherment
```
Add the following to the [ CA_default ] section of openssl-ca.cnf.
```
base_dir      = .
certificate   = $base_dir/cacert.pem   # The CA certifcate
private_key   = $base_dir/cakey.pem    # The CA private key
new_certs_dir = $base_dir              # Location for new certs after signing
database      = $base_dir/index.txt    # Database index file
serial        = $base_dir/serial.txt   # The current serial number

unique_subject = no  # Set to 'no' to allow creation of
                     # several certificates with same subject.

```
Third, touch index.txt and serial.txt:
```
$ touch index.txt
$ echo '01' > serial.txt
```
Then, perform the following:
```
$ openssl ca -config openssl-ca.cnf -policy signing_policy -extensions signing_req -out servercert.pem -infiles servercert.csr
```

### Copying the Certificates
1. Copy the servercert.pem, serverkey.pem and cacert.pem file to /etc/mosquitto/certs
2. Mention the path of the certs in the configuration file of mosquitto
3. Copy the cacert.pem file in /var/edge-orchestration/mqtt/certs folder in client side
4. Mosquitto broker allows to generate password for the users you want to allow connection.For example password can be generated for User A in a file named A.txt and path can be specified in the configuration file. Similarly the password can be generated for another user say User C as C.txt with password saved in it.
** By Default allow_anonymous has to be kept true



# 5. Configurations for MQTT broker
MQTT broker provides the following secure mode
1. TLS Handshake to secure the channel of communication
2. Creating passwords for users whom we want the communication to be allowed for the broker thereby specifying the authorization
So mosquitto allows one to store the password file in any location and provide the path in mosquitto conf file.

### 3 level of Security between MQTT broker and client communication

    1. Allow_anonymous = true (No security, port 1883)
    2. Allow_anonymous = true (With TLS Certificates making only channel of communication secure but allow any user to communicate who has the certs,port 8883)
    3. Allow_anonymous = false (With TLS certificate and password for specific users thereby providing authentication for communication,port 8883)


***Please Note :*** By Default the allow_anonymous = true should be mentioned in the configuration file. As the security level is increased the value can be changed and passwords can be included

# 6. Run Cloudsync 

The edge orchestration is build and run using following command with option CLOUD_SYNC set to true
```
docker run -it -d --privileged --network="host" --name edge-orchestration -e CLOUD_SYNC=true -v /var/edge-orchestration/:/var/edge-orchestration/:rw -v /var/run/docker.sock:/var/run/docker.sock:rw -v /proc/:/process/:ro  lfedge/edge-home-orchestration-go:latest
```
From the another terminal/post make a curl command as follows to publish data using home edge to the broker running on AWS endpoint

```
curl --location --request POST 'http://<ip where edgeorchestration is running>:56001/api/v1/orchestration/cloudsyncmgr/publish' \
--header 'Content-Type: text/plain' \
--data-raw '{
    "appid": "<appid of service app>",
    "payload": "{Another data from TV1 and testdata}",
    "topic": "home1/livingroom",
    "url" : "<AWS public IP>"
}'
```