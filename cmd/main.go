package main

import (
	"context"
	"fmt"
	"github.com/dotbitHQ/docker-events-monitor/config"
	"github.com/dotbitHQ/docker-events-monitor/shell"
	"github.com/dotbitHQ/docker-events-monitor/supervisor"
	"github.com/dotbitHQ/docker-events-monitor/tool"
	"github.com/scorpiotzh/mylog"
	"github.com/scorpiotzh/toolib"
	"github.com/urfave/cli/v2"
	"os"
	"sync"
	"time"
)

var (
	log               = tool.GetLog("main", mylog.LevelDebug)
	exit              = make(chan struct{})
	ctxServer, cancel = context.WithCancel(context.Background())
	wgServer          = sync.WaitGroup{}
)

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Load configuration from `FILE`",
			},
		},
		Action: runServer,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func runServer(ctx *cli.Context) error {
	// config file
	configFilePath := ctx.String("config")
	if err := config.InitCfg(configFilePath); err != nil {
		return err
	}

	// config file watcher
	watcher, err := config.AddCfgFileWatcher(configFilePath)
	if err != nil {
		return fmt.Errorf("AddCfgFileWatcher err: %s", err.Error())
	}

	// ============= service start =============

	var deCmd shell.DockerEventsCmd
	deCmd.Ctx = ctxServer
	deCmd.Wg = &wgServer
	deCmd.InitContainerMap(config.Cfg.Server.Containers, config.Cfg.Server.Status)
	deCmd.Exec()

	var el supervisor.EventsListener
	el.Ctx = ctxServer
	el.Wg = &wgServer
	el.Run()

	// ============= service end =============
	toolib.ExitMonitoring(func(sig os.Signal) {
		log.Warn("ExitMonitoring:", sig.String())
		if watcher != nil {
			log.Warn("close watcher ... ")
			_ = watcher.Close()
		}

		cancel()
		el.Closed()
		wgServer.Wait()
		log.Warn("success exit server. bye bye!")
		time.Sleep(time.Second)
		exit <- struct{}{}
	})

	<-exit

	return nil
}
