package main

import (
	"os"

	"./gdrive2discord"
)

var version = "0.0.1a"

func main() {
	logger := gdrive2discord.NewLogger(os.Stdout, "", 0)
	logger.Info("gdrive2discord version:%s", version)
	if len(os.Args) != 2 {
		logger.Error("usage: %s <configuration_file>", os.Args[0])
		os.Exit(1)
	}

	configuration, err := gdrive2discord.LoadConfiguration(os.Args[1])
	if err != nil {
		logger.Error("cannot read configuration: %s", err)
		os.Exit(1)
	}
	env := gdrive2discord.NewEnvironment(version, configuration, logger)

	go gdrive2discord.EventLoop(env)
	gdrive2discord.ServeHttp(env)
}
