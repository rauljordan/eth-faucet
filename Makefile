go:
	go get . && go build -o ./dist/faucet .
protos:
	protoc -Iproto --grpc-gateway_out=logtostderr=true,paths=source_relative:proto\
	  proto/faucet/faucet.proto
	protoc -Iproto --go_out=proto --go_opt=paths=source_relative \
    --go-grpc_out=proto --go-grpc_opt=paths=source_relative \
    proto/faucet/faucet.proto
release:
	env GOOS=darwin GOARCH=amd64 go build -o ./dist/faucet-v1.0.1-darwin-amd64 .
	env GOOS=windows GOARCH=amd64 go build -o ./dist/faucet-v1.0.1-windows-amd64 .
	env GOOS=linux GOARCH=amd64 go build -o ./dist/faucet-v1.0.1-linux-amd64 .
	env GOOS=linux GOARCH=arm64 go build -o ./dist/faucet-v1.0.1-linux-amd64 .
	shasum ./dist/faucet-v1.0.1-windows-amd64 > ./dist/faucet-v1.0.1-darwin-amd64.sha256
	shasum ./dist/faucet-v1.0.1-darwin-amd64 > ./dist/faucet-v1.0.1-darwin-amd64.sha256
	shasum ./dist/faucet-v1.0.1-linux-amd64 > ./dist/faucet-v1.0.1-darwin-amd64.sha256
	shasum ./dist/faucet-v1.0.1-darwin-arm64 > ./dist/faucet-v1.0.1-darwin-arm64.sha256
