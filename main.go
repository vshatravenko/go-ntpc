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
	"github.com/vshatravenko/go-ntpc/pkg/systime"
)

func main() {
	app := &cli.App{
		Name:        "ntpc",
		Description: "Connect to a given NTP server and calculate latency metrics",
		Commands: []*cli.Command{
			{
				Name:        "connect",
				Description: "Exchange two NTP packets with a server and print the clock offset and adjusted time",
				Action:      connectAction,
			},
			{
				Name:        "update-systime",
				Description: "Poll an NTP server, adjusting system time until the clock offset is minimized (dry-run by default)",
				Action:      updateSystimeAction,
			},
		},
		DefaultCommand: "connect",
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("Ouch, got an error: %s\n", err.Error())
		os.Exit(1)
	}
}

func connectAction(ctx *cli.Context) error {
	client, conf, err := setup()
	if err != nil {
		return fmt.Errorf("could not set up the NTP client: %v", err)
	}

	fmt.Printf("Connecting to %s:%d\n", conf.RemoteHost, conf.RemotePort)
	res, err := client.Exchange()
	if err != nil {
		return fmt.Errorf("NTP exchange failed: %v", err)
	}

	fmt.Printf("Adjusted time: %s, offset: %s\n", time.Now().Add(res.ClockOffset), res.ClockOffset)

	return nil
}

func updateSystimeAction(ctx *cli.Context) error {
	client, conf, err := setup()
	if err != nil {
		return fmt.Errorf("could not set up the NTP client: %v", err)
	}

	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ticker.C:
			res, err := client.Exchange()
			if err != nil {
				fmt.Printf("err: could not exchange NTP packets: %s\n", err)
				continue
			}

			fmt.Printf("current clock offset: %s\n", res.ClockOffset)

			if abs(int64(res.ClockOffset)) < int64(5*time.Millisecond) {
				fmt.Println("optimal offset reached, exiting!")
				return nil
			}

			err = systime.UpdateSysTime(int64(res.ClockOffset), conf.SystimeUpdateEnabled)
			if err != nil {
				return fmt.Errorf("system time update failed: %s", err)
			}
		}
	}
}

func setup() (*ntp.Client, *config.Config, error) {
	conf := &config.Config{}

	if err := env.Parse(conf); err != nil {
		return nil, nil, fmt.Errorf("could not parse config: %+v", err)
	}

	addr, err := ntp.SetupRemoteAddr(conf.RemoteHost, conf.RemotePort)
	if err != nil {
		return nil, nil, fmt.Errorf("could not set up remote address: %v", err)
	}

	client, err := ntp.NewClient(addr)
	if err != nil {
		return nil, nil, fmt.Errorf("could not set up the client: %v", err)
	}

	return client, conf, nil
}

func abs(a int64) int64 {
	if a < 0 {
		return -a
	}

	return a
}
