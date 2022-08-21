## learnTLS_go
Implementación tls para gRPC Server + gRPC Gateway Server. Inspirado en el articulo https://medium.com/@mertkimyonsen/securing-grpc-connection-with-ssl-tls-certificate-using-go-db3852fe89dd 

Gen certs
```makefile
make gen-cert
```

Gen stubs
```zsh
buf generate
```

Test with grpcurl - gRPC Server

```zsh
./grpcurl -insecure -d '{"name":"Juan"}' localhost:8080 helloworld.Greete│
r.SayHello                                                                          
{                                                                                   
  "message": "Juan world"                                                           
}  
```

Test with curl - gRPC Gateway 

```zsh
curl -k -X POST https://localhost:8080/v1/example/echo -d '{"name":"juan"}'
# response
{"message":"juan world"}
```



