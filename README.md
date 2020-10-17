# Ethereum Goerli Testnet Faucet

This project defines a production ready faucet for the Ethereum Goerli test network, allowing users to receive 32.5 Goerli ETH every 24 hours to an address from a max of 5 different IP addresses (configurable). The API tracks IP addresses and wallet addresses which requested and flushes them out every 24 hours to allow them to request again the next day.

## Running the Faucet API

Install Go [here](https://golang.org/doc/install)

Available flags are as follows:

- -rpc-port, default: 5000, port to serve gRPC requests
- -host, default: "127.0.0.1", host address to serve gRPC requests
- -gateway-port, default: 8000, port to serve JSON-RPC requests
- -gateway-host, default: 127.0.0.1, host address to serve JSON-RPC requests
- -allowed-origins, default: "*", allowed origins for JSON-RPC requests, comma-separated
- -captcha-host, default: "", host for the captcha
- -captcha-secret, default: "", secret to verify recaptcha
- -rpc, default: "", RPC address of a running geth node
- -private-key, default: "", the private key of funder
- -min-score, default: 0.9, minimum captcha score

Build the API:

```
make go
```

Run the API server:

```
./main \
  -rpc=<GETH_RPC> \
  -private-key=<GOERLI_PRIVATE_KEY_HEX> \
  -captcha-host=<CAPTCHA_HOST> \
  -captcha-secret=<CAPTCHA_SECRET>
```

You will have a JSON-RPC endpoint running on 127.0.0.1:8000 by default and a gRPC server on 127.0.0.1:5000, which can be accessed from your frontend.

## Running the Example Angular Application

```
cd web
npm install
npm start
```

Then navigate to http://localhost:4200 and try it out!
