package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/urfave/cli/v2"
	"github.com/vshatravenko/go-ntpc/pkg/config"
	"github.com/vshatravenko/go-ntpc/pkg/ntp"
)

func main() {
	app := &cli.App{
		Name:   "ntpc",
		Usage:  "Connect to a given NTP server and calculate latency metrics",
		Action: connectAction,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("Ouch, got an error: %s\n", err.Error())
		os.Exit(1)
	}
}

func connectAction(ctx *cli.Context) error {
	conf := config.Config{}

	if err := env.Parse(&conf); err != nil {
		return fmt.Errorf("could not parse config: %+v", err)
	}

	addr, err := ntp.SetupRemoteAddr(conf.RemoteHost, conf.RemotePort)
	if err != nil {
		return fmt.Errorf("could not set up remote address: %v", err)
	}

	fmt.Printf("Connecting to %s:%d (%s)\n", conf.RemoteHost, conf.RemotePort, addr)

	client, err := ntp.NewClient(addr)
	if err != nil {
		return fmt.Errorf("could not set up the client: %v", err)
	}

	res, err := client.Exchange()
	if err != nil {
		return fmt.Errorf("NTP exchange failed: %v", err)
	}

	fmt.Printf("Adjusted time: %s, offset: %s\n", time.Now().Add(res.ClockOffset), res.ClockOffset)

	return nil
}
