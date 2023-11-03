# >>>>>>>>>>>>>>>>>> Root Cert <<<<<<<<<<<<<<<<<<<<<<
# ca.key
openssl genrsa -out ca.key 2048

# ca.crt
openssl req -new -key ca.key -x509 -days 3650 -out ca.crt -subj /C=CN/ST=Shaanxi/L="Xi'an"/O=RedHat/CN="Xi'an Redhat Root"

# >>>>>>>>>>>>>>>>>> Server <<<<<<<<<<<<<<<<<<<<<<
# server.key
openssl genrsa -out server.key 2048

# server.csr
openssl req -new -nodes -key server.key -out server.csr -subj /C=CN/ST="Shaanxi"/L="Xi'an"/O="RedHat"/CN="server"

# server.crt
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt

# >>>>>>>>>>>>>>>>>> Client <<<<<<<<<<<<<<<<<<<<<<
# client.key
openssl genrsa -out client.key 2048

# client.csr
openssl req -new -nodes -key client.key -out client.csr -subj /C=CN/ST="Shaanxi"/L="Xi'an"/O="RedHat"/CN="client"

# client.crt
openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client.crt