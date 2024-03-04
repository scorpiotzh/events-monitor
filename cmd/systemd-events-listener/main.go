package main

import (
	"events-monitor/systemd"
	"fmt"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "key",
				Aliases: []string{"k"},
				Usage:   "Lark notify key",
			},
			&cli.MultiStringFlag{
				Target: &cli.StringSliceFlag{
					Name:    "services",
					Aliases: []string{"s"},
					Usage:   "list of systemd service name",
				},
			},
		},
		Action: runServer,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func runServer(ctx *cli.Context) error {
	key := ctx.String("key")
	if key == "" {
		return fmt.Errorf("key is nil")
	}
	services := ctx.StringSlice("services")
	if len(services) == 0 {
		return fmt.Errorf("no ststemd service set")
	}
	l := systemd.EventListener{LarkKey: key, Services: services}
	return l.Run()
}
