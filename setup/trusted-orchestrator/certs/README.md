# About These Certificates

The following steps were taken to create these public/private certificate pairs and
are not the proper way to secure communication in Vault. This is for demo purposes only.

Generate the private key with :
```bash
openssl genrsa -out private.pem 2048
```

Generate the public key:
```bash
openssl rsa -in private.pem -outform PEM -pubout -out public.pem
```

Create a Certificate Signing Request using our private key:
```bash
openssl req -new -key private.pem -out certificate.csr
```

Create a self-signed certificate using the previously created resources (good for 10 years):
```bash
openssl x509 -req -days 3650 -in certificate.csr -signkey private.pem -out certificate.crt
```