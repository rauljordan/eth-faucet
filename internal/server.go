package internal

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"net"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/prestonvanloon/go-recaptcha"
	faucetpb "github.com/rauljordan/eth-faucet/proto/faucet"
	gateway "github.com/rauljordan/minimal-grpc-gateway"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	log = logrus.WithField("prefix", "api")
)

var (
	funded        = make(map[string]bool)
	ipCounter     = make(map[string]int)
	fundingLock   sync.Mutex
	pruneDuration = time.Hour * 4
	_             = faucetpb.FaucetServer(&Server{})
)

// Config for the faucet server.
type Config struct {
	GrpcPort          int      `mapstructure:"grpc-port"`
	GrpcHost          string   `mapstructure:"grpc-host"`
	HttpPort          int      `mapstructure:"http-port"`
	HttpHost          string   `mapstructure:"http-host"`
	AllowedOrigins    []string `mapstructure:"allowed-origins"`
	CaptchaHost       string   `mapstructure:"captcha-host"`
	CaptchaSecret     string   `mapstructure:"captcha-secret"`
	CaptchaMinScore   float64  `mapstructure:"captcha-min-score"`
	Web3Provider      string   `mapstructure:"web3-provider"`
	PrivateKey        string   `mapstructure:"private-key"`
	FundingAmount     string   `mapstructure:"funding-amount"`
	GasLimit          string   `mapstructure:"gas-limit"`
	IpLimitPerAddress int      `mapstructure:"ip-limit-per-address"`
}

// Server capable of funding requests for faucet ETH via gRPC and REST HTTP.
type Server struct {
	faucetpb.UnimplementedFaucetServer
	cfg           *Config
	captcha       recaptcha.Recaptcha
	client        *ethclient.Client
	funder        common.Address
	pk            *ecdsa.PrivateKey
	minScore      float64
	captchaHost   string
	fundingAmount *big.Int
}

func NewServer(cfg *Config) (*Server, error) {
	privKeyHex := cfg.PrivateKey
	if strings.HasPrefix(privKeyHex, "0x") {
		privKeyHex = privKeyHex[2:]
	}
	pk, err := crypto.HexToECDSA(privKeyHex)
	if err != nil {
		return nil, fmt.Errorf("could not parse funder private key: %v", err)
	}
	fundingAmount, ok := new(big.Int).SetString(cfg.FundingAmount, 10)
	if !ok {
		return nil, errors.New("could not set funding amount")
	}
	return &Server{
		cfg:           cfg,
		captcha:       recaptcha.Recaptcha{RecaptchaPrivateKey: cfg.CaptchaSecret},
		pk:            pk,
		fundingAmount: fundingAmount,
	}, nil
}

func (s *Server) Start() {
	ctx := context.Background()
	runtime.GOMAXPROCS(runtime.NumCPU())

	s.queryFundsLeft()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.cfg.GrpcPort))
	if err != nil {
		log.WithError(err).Fatalf("Could not listen on port %d", s.cfg.GrpcPort)
	}
	grpcServer := grpc.NewServer()

	faucetpb.RegisterFaucetServer(grpcServer, s)
	reflection.Register(grpcServer)

	// Check IP addresses and reset their max request count over time.
	go ipAddressCounterWatcher()

	// Start a gRPC server.
	go func() {
		log.Infof("Serving gRPC requests on port %s:%d", s.cfg.GrpcHost, s.cfg.GrpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.WithError(err).Fatal("Stopped server")
		}
	}()

	// Start a gRPC Gateway to serve JSON-HTTP requests.
	gatewaySrv := gateway.New(ctx, &gateway.Config{
		GatewayAddress:      fmt.Sprintf("%s:%d", s.cfg.HttpHost, s.cfg.HttpPort),
		RemoteAddress:       fmt.Sprintf("%s:%d", s.cfg.GrpcHost, s.cfg.GrpcPort),
		AllowedOrigins:      s.cfg.AllowedOrigins,
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

func (s *Server) queryFundsLeft() {
	client, err := ethclient.DialContext(context.Background(), s.cfg.Web3Provider)
	if err != nil {
		log.WithError(err).Fatalf("Could not dial %s", s.cfg.Web3Provider)
	}
	funder := crypto.PubkeyToAddress(s.pk.PublicKey)
	bal, err := client.BalanceAt(context.Background(), funder, nil)
	if err != nil {
		log.WithError(err).Fatalf("Could not retrieve funder's current balance")
	}

	log.WithFields(logrus.Fields{
		"fundsInWei": bal,
		"publicKey":  funder.Hex(),
	}).Info("Funder details")
}

// Reduce the counter for each ip every few hours.
func ipAddressCounterWatcher() {
	ticker := time.NewTicker(pruneDuration)
	for {
		<-ticker.C
		fundingLock.Lock()
		log.Info("Decreasing requests counter for all recorded IP addresses")
		for ip, ctr := range ipCounter {
			if ctr == 0 {
				continue
			}
			ipCounter[ip] = ctr - 1
		}
		fundingLock.Unlock()
	}
}
