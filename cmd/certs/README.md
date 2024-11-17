# Certifaction

1. 生成 CA（根证书）：

```shell
# 生成 CA 私钥
openssl genrsa -out ca.key 2048

# 生成 CA 根证书
openssl req -x509 -new -nodes -key ca.key -subj "/CN=MyCA" -days 3650 -out ca.crt
```

2. 生成服务端证书

```shell
# 生成服务端私钥
openssl genrsa -out server.key 2048

# 生成服务端证书签名请求（CSR）
openssl req -new -key server.key -subj "/CN=server.local" -out server.csr

# 使用 CA 签署服务端证书
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 365 -sha256
```

3. 生成客户端证书

```shell
# 生成客户端私钥
openssl genrsa -out client.key 2048

# 生成客户端证书签名请求（CSR）
openssl req -new -key client.key -subj "/CN=client.local" -out client.csr

# 使用 CA 签署客户端证书
openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client.crt -days 365 -sha256

```

4. curl 

```shell
curl -v --cert ./certs/client.crt --key ./certs/client.key --cacert ./certs/ca.crt https://localhost
```
