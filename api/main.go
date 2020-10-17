package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	"github.com/prestonvanloon/go-recaptcha"
	faucetpb "github.com/rauljordan/goerli-faucet/api/proto/faucet"
	gateway "github.com/rauljordan/minimal-grpc-gateway"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	port           = flag.Int("grpc-port", 5000, "Port to serve gRPC requests")
	host           = flag.String("grpc-host", "127.0.0.1", "Host address to serve gRPC requests")
	gatewayPort    = flag.Int("gateway-port", 8000, "Port to serve JSON-RPC requests")
	gatewayHost    = flag.String("gateway-host", "127.0.0.1", "Host address to serve JSON-RPC requests")
	allowedOrigins = flag.String("allowed-origins", "*", "Allowed origins for JSON-RPC requests, comma-separated")
	captchaHost    = flag.String("captcha-host", "", "Host for the captcha")
	captchaSecret  = flag.String("captcha-secret", "", "Secret to verify recaptcha")
	rpcPath        = flag.String("rpc", "", "RPC address of a running geth node")
	privateKey     = flag.String("private-key", "", "The private key of funder")
	minScore       = flag.Float64("min-score", 0.9, "Minimum captcha score")
	log            = logrus.WithField("prefix", "api")
)

func main() {
	flag.Parse()
	if *captchaHost == "" {
		log.Fatalf("-captcha-host required (ex: prylabs.net)")
	}
	if *captchaSecret == "" {
		log.Fatalf("-captcha-secret required")
	}
	if *privateKey == "" {
		log.Fatalf("-private-key hex string for a goerli address required")
	}
	if *rpcPath == "" {
		log.Fatalf("-rpc http or ipc endpoint to an eth1 goerli node required (ex: http://localhost:8545)")
	}
	if *allowedOrigins == "*" {
		log.Warn(
			"Allowing origin requests from any source '*', use the -allowed-origins flag to customize, ex: http://localhost:4200",
		)
	}

	ctx := context.Background()
	runtime.GOMAXPROCS(runtime.NumCPU())
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.WithError(err).Fatalf("Could not listen on port %d", *port)
	}
	grpcServer := grpc.NewServer()
	server, err := New(
		recaptcha.Recaptcha{RecaptchaPrivateKey: *captchaSecret},
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

	// Check IP addresses and reset their max request count over time.
	go ipAddressCounterWatcher()

	// Start a gRPC server.
	go func() {
		log.Infof("Serving gRPC requests on port %d\n", *port)
		if err := grpcServer.Serve(lis); err != nil {
			log.WithError(err).Fatal("Stopped server")
		}
	}()

	// Start a gRPC Gateway to serve JSON-HTTP requests.
	origins := []string{*allowedOrigins}
	if strings.Contains(*allowedOrigins, ",") {
		origins = strings.Split(*allowedOrigins, ",")
	}
	gatewaySrv := gateway.New(ctx, &gateway.Config{
		GatewayAddress:      fmt.Sprintf("%s:%d", *gatewayHost, *gatewayPort),
		RemoteAddress:       fmt.Sprintf("%s:%d", *host, *port),
		AllowedOrigins:      origins,
		EndpointsToRegister: []gateway.RegistrationFunc{faucetpb.RegisterFaucetHandlerFromEndpoint},
	})
	gatewaySrv.Start()

	// Listen for any process interrupts.
	stop := make(chan struct{})
	go func() {
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(sigc)
		<-sigc
		logrus.Info("Got interrupt, shutting down...")
		grpcServer.GracefulStop()
		stop <- struct{}{}
	}()

	// Wait for stop channel to be closed.
	<-stop
}
