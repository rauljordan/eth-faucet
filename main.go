package main

import (
	"github.com/rauljordan/eth-faucet/cmd"
	"github.com/sirupsen/logrus"
)

func main() {
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	logrus.SetFormatter(customFormatter)
	customFormatter.FullTimestamp = true
	cmd.Execute()
}
