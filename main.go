package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	helloworldpb "github.com/juanmachuca95/learnTLS_go/proto/helloworld"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/reflection"
)

func init() {
	logger := grpclog.NewLoggerV2(os.Stdout, ioutil.Discard, ioutil.Discard)
	grpclog.SetLoggerV2(logger)
}

type server struct {
	helloworldpb.UnimplementedGreeterServer
}

func NewServer() *server {
	return &server{}
}

func (s *server) SayHello(ctx context.Context, in *helloworldpb.HelloRequest) (*helloworldpb.HelloReply, error) {
	return &helloworldpb.HelloReply{Message: in.Name + " world"}, nil
}

// Ejemplo to implement tls on go server GRPC + REST
func main() {
	// 1.read ca's cert, verify to client's certificate
	caPem, err := ioutil.ReadFile("cert/ca-cert.pem")
	if err != nil {
		log.Fatal("readfile ", err)
	}

	// 2. create cert pool and append ca's cert
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caPem) {
		log.Fatal("append to certpool ", err)
	}

	// 3. read server cert & key
	serverCert, err := tls.LoadX509KeyPair("cert/server-cert.pem", "cert/server-key.pem")
	if err != nil {
		log.Fatal("load x509", err)
	}

	// 5. configuration of the certificate what we want to
	conf := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.NoClientCert,
		ClientCAs:    certPool,
	}

	// 6. create tls credentials
	tlsCredentials := credentials.NewTLS(conf)
	// Create a gRPC Server object
	grpcServer := grpc.NewServer(grpc.Creds(tlsCredentials))

	helloworldpb.RegisterGreeterServer(grpcServer, &server{})

	// Rest Server
	mux := runtime.NewServeMux()
	err = helloworldpb.RegisterGreeterHandlerServer(context.Background(), mux, &server{})
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	// enable reflection
	reflection.Register(grpcServer)

	err = http.ListenAndServeTLS(":8080", "cert/server-cert.pem", "cert/server-key.pem", grpcHandlerFunc(grpcServer, mux))
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}

func grpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	return h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			log.Println("GRPC")
			grpcServer.ServeHTTP(w, r)
		} else {
			log.Println("REST")
			otherHandler.ServeHTTP(w, r)
		}
	}), &http2.Server{})
}
