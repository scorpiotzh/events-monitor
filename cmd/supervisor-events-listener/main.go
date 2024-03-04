package main

import (
	"fmt"
	"github.com/scorpiotzh/docker-events-monitor/supervisor"
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
				Usage:   "Load key",
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
	//fmt.Println(key)
	var el supervisor.EventsListener
	el.Key = key
	el.Run()
	return nil
}
