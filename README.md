# Ethereum Testnet Faucet 🚿

This project defines a production ready faucet for the Ethereum test networks, allowing users to request and receive a specified amount of ETH every 24 hours to an address from a max of N different IP addresses (configurable) after passing [Google recaptcha](https://www.google.com/recaptcha/about/) verification. The API tracks IP addresses and wallet addresses which requested and resets them at configurable intervals.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT) [![Go](https://github.com/rauljordan/eth-faucet/actions/workflows/go.yml/badge.svg?branch=master)](https://github.com/rauljordan/eth-faucet/actions/workflows/go.yml) ![ReportCard](https://goreportcard.com/badge/github.com/rauljordan/eth-faucet)

<img width="400" alt="Screen Shot 2021-04-12 at 9 44 16 PM" src="https://user-images.githubusercontent.com/5572669/114489164-4a7ed780-9bd8-11eb-92ee-3e140a25fafd.png">


## Links

- See [releases](https://github.com/rauljordan/eth-faucet/releases) for a detailed version history.
- To contribute, see [here](#contributing)
- Please [open an issue](https://github.com/rauljordan/eth-faucet/issues/new) if anything is missing, unclear, or for any suggested improvements

## Installation

Download the latest version of [Go](https://golang.org/dl/). Then, build the project.

```bash
make go
```

## Usage

### Faucet Server

```bash
./dist/faucet --help
```

Outputs

```
Run a faucet server for Ethereum using captcha for verification and ip-based rate-limiting

Usage:
  faucet [flags]
```

1. Sign-up for Google recaptcha in the admin portal [here](http://www.google.com/recaptcha/admin)
2. Make note of the captcha secret key and the captcha site key
3. Obtain your testnet private key where the faucet funds will be coming from in *hex string format* (such as 0xa8b01...)
4. Obtain the address of your Ethereum node's HTTP API endpoint (by default, the faucet server uses http://localhost:8545 as the web3-provider)
5. Run the faucet server with the required [parameters](#parameters)

The faucet hosts an http JSON API on `localhost:8000` by default and a gRPC server on `localhost:5000` for client access. Further customizations and required parameters are specified below:

#### Parameters

The following are the available parameters to the faucet server:

**Required Flags**

| flag   | Description                                 | Default Value
| ------ | ------------------------------------------- | ------------- |
| --web3-provider | HTTP web3provider endpoint to an Ethereum node | "http://localhost:8545" | Yes
| --captcha-host |  Host for the captcha validation    | "" 
| --captcha-secret | Secret for Google captcha validation | ""
| --private-key | Private key hex string of the funding account | ""

**Web Server Flags**

| flag   | Description                                 | Default Value
| ------ | ------------------------------------------- | -------------
| --http-host | Host to serve REST http requests | 127.0.0.1
| --http-port | Port to serve REST http requests | 8000
| --grpc-host | Host to serve gRPC requests | 127.0.0.1
| --grpc-port | Port to serve gRPC requests | 5000
| --allowed-origins | Comma-separated list of allowed origins | "*"

**Misc. Flags**

| flag   | Description                                 | Default Value
| ------ | ------------------------------------------- | -------------
| --config | Path to yaml configuration file for flags | ""
| --captcha-min-score | Minimum passing captcha score | 0.9
| --chain-id | Chain id of the Ethereum network used | 5 (Goerli)
| --funding-amount | Amount in wei to fund with each request | 32500000000000000000
| --gas-limit | Gas limit for funding transactions | 40000
| --ip-limit-per-address | Number of ip's allowed per funding address | 5


#### Configuration

You can configure the faucet by using a yaml configuration file instead of command-line flags as follows:

```yaml
chain-id: 9999
http-port: 8080
# Insert all other desired customizations below...
```

and running the faucet server by specifying the path to the configuration file as follows:

```
./dist/faucet --config=/path/to/config.yaml
```

### Sample Angular Project

1. Install the latest version of [Node.js](https://nodejs.org/en/download/). 
2. Paste in the **captcha site key** from the [usage](#usage) step in `./web/ng/src/environments/environment.ts` or `./web/ng/src/environments/environment.prod.ts` if building for production
3. Run the faucet server as specified in [usage](#usage)
4. Run the Angular app as follows:

```
cd web/ng
npm install
npm start
```

Then navigate to http://localhost:4200 and try it out!

You can build it for production with `npm run build`

## Contributing

Regenerating protobufs:

1. Install the latest version of [`protoc`](https://grpc.io/docs/protoc-installation/)
2. `make proto`

Running tests:

```
go test ./... -v
```

## License

The project is licensed under the MIT License.
