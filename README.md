# grpc_demo
grpc学习

## TLS证书生成

### 私钥

```shell	
openssl ecparam -genkey -name secp384r1 -out server.key
```

### 自签公钥

```shell
openssl req -new -x509 -sha256 -key server.key -out server.pem -days 3650
```

填写信息：

```shell
You are about to be asked to enter information that will be incorporated
into your certificate request.
What you are about to enter is what is called a Distinguished Name or a DN.
There are quite a few fields but you can leave some blank
For some fields there will be a default value,
If you enter '.', the field will be left blank.
-----
Country Name (2 letter code) [AU]:CN 
State or Province Name (full name) [Some-State]:
Locality Name (eg, city) []:
Organization Name (eg, company) [Internet Widgits Pty Ltd]:
Organizational Unit Name (eg, section) []:
Common Name (e.g. server FQDN or YOUR name) []:grpc_demo
Email Address []:
```

## 基于CA的TLS证书认证

### 根证书

#### 生成key

```shell	
openssl genrsa -out ca.key 2048
```

#### 生成密钥

```shell
openssl req -new -x509 -days 7200 -key ca.key -out ca.pem
```

填写信息：

```shell
You are about to be asked to enter information that will be incorporated
into your certificate request.
What you are about to enter is what is called a Distinguished Name or a DN.
There are quite a few fields but you can leave some blank
For some fields there will be a default value,
If you enter '.', the field will be left blank.
-----
Country Name (2 letter code) [AU]:CN
State or Province Name (full name) [Some-State]:
Locality Name (eg, city) []:
Organization Name (eg, company) [Internet Widgits Pty Ltd]:
Organizational Unit Name (eg, section) []:
Common Name (e.g. server FQDN or YOUR name) []:grpc_demo
Email Address []:
```

### Server

#### 生成key

```shell
openssl ecparam -genkey -name secp384r1 -out server.key
```

#### 生成csr

CSR 是 Cerificate Signing Request 的英文缩写，为证书请求文件。主要作用是 CA 会利用 CSR 文件进行签名使得攻击者无法伪装或篡改原有证书

```shell	
openssl req -new -key server.key -out server.csr
```

填写信息：

```shell
You are about to be asked to enter information that will be incorporated
into your certificate request.
What you are about to enter is what is called a Distinguished Name or a DN.
There are quite a few fields but you can leave some blank
For some fields there will be a default value,
If you enter '.', the field will be left blank.
-----
Country Name (2 letter code) [AU]:CN
State or Province Name (full name) [Some-State]:
Locality Name (eg, city) []:
Organization Name (eg, company) [Internet Widgits Pty Ltd]:
Organizational Unit Name (eg, section) []:
Common Name (e.g. server FQDN or YOUR name) []:grpc_demo 
Email Address []:

Please enter the following 'extra' attributes
to be sent with your certificate request
A challenge password []:aiwdqu82e1ie0-12^[[19~2iasii
string is too long, it needs to be no more than 20 bytes long
A challenge password []:821wsd88*(29sjjmxz
An optional company name []:
```

#### 基于CA签发

```shell	
openssl x509 -req -sha256 -CA ca.pem -CAkey ca.key -CAcreateserial -days 3650 -in server.csr -out server.pem 
================= output ====================
Signature ok
subject=C = CN, ST = Some-State, O = Internet Widgits Pty Ltd, CN = grpc_demo
Getting CA Private Key
```

### Client

#### 生成key

```shell	
openssl ecparam -genkey -name secp384r1 -out client.key
```

#### 生成CSR

```shell
openssl req -new -key client.key -out client.csr
```

#### 基于CA签发

```shell
openssl x509 -req -sha256 -CA ca.pem -CAkey ca.key -CAcreateserial -days 3650 -in client.csr -out client.pem
```

