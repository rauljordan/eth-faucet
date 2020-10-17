package main

import (
	"flag"
	"fmt"
	"net"
	"runtime"

	"github.com/prestonvanloon/go-recaptcha"
	faucetpb "github.com/rauljordan/goerli-faucet/api/proto/faucet"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	port            = flag.Int("port", 8000, "Port to server gRPC service")
	captchaHost     = flag.String("captcha-host", "", "Host for the captcha")
	recaptchaSecret = flag.String("recaptcha-secret", "", "Secret to verify recaptcha")
	rpcPath         = flag.String("rpc", "", "RPC address of a running geth node")
	privateKey      = flag.String("private-key", "", "The private key of funder")
	minScore        = flag.Float64("min-score", 0.9, "Minimum captcha score")
	log             = logrus.WithField("prefix", "api")
)

func main() {
	flag.Parse()

	runtime.GOMAXPROCS(runtime.NumCPU())
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.WithError(err).Fatal("Could not listen on port %d", *port)
	}
	grpcServer := grpc.NewServer()
	server, err := New(
		recaptcha.Recaptcha{RecaptchaPrivateKey: *recaptchaSecret},
		*rpcPath,
		*privateKey,
		*minScore,
		*captchaHost,
	)
	if err != nil {
		log.WithError(err).Fatal("Could not initialize faucet server")
	}

	faucetpb.RegisterFaucetServer(grpcServer, server)
	reflection.Register(grpcServer)
	go ipAddressCounterWatcher()

	log.Infof("Serving gRPC requests on port %d\n", *port)
	if err := grpcServer.Serve(lis); err != nil {
		log.WithError(err).Fatal("Stopped server")
	}
}
