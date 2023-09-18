package shell

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/scorpiotzh/docker-events-monitor/config"
	"github.com/scorpiotzh/docker-events-monitor/notify"
	"github.com/scorpiotzh/docker-events-monitor/tool"
	"github.com/scorpiotzh/mylog"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"
)

var (
	log = tool.GetLog("shell", mylog.LevelDebug)
)

type DockerEventsCmd struct {
	Ctx          context.Context
	Wg           *sync.WaitGroup
	ContainerMap map[string]struct{}
	StatusMap    map[string]struct{}
	StdoutReader *bufio.Reader
}

type ContainerInfo struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Time   int64  `json:"time"`
}

func (c *DockerEventsCmd) InitContainerMap(containers, status []string) {
	c.ContainerMap = make(map[string]struct{})
	for _, v := range containers {
		c.ContainerMap[strings.ToLower(v)] = struct{}{}
	}
	c.StatusMap = make(map[string]struct{})
	for _, v := range status {
		c.StatusMap[strings.ToLower(v)] = struct{}{}
	}
}

func (c *DockerEventsCmd) Exec() {
	go func() {
		if err := c.exec(); err != nil {
			log.Error("Exec err: %s", err.Error())
		}
	}()
}

func (c *DockerEventsCmd) exec() error {
	command := `docker events --filter 'type=container' --format '{"status":"{{.Status}}","name":"{{.Actor.Attributes.name}}","time":{{.Time}}}'`
	cmd := exec.CommandContext(c.Ctx, "bash", "-c", command)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("StdoutPipe err: %s", err.Error())
	}
	c.StdoutReader = bufio.NewReader(stdout)
	c.runEventsFormat()

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cmd.Run() err: %s", err.Error())
	}
	return nil
}

func (c *DockerEventsCmd) runEventsFormat() {
	title := fmt.Sprintf("Docker 服务监控(%s)", config.Cfg.Server.Name)
	c.Wg.Add(1)
	go func() {
		for {
			select {
			case <-c.Ctx.Done():
				c.Wg.Done()
				log.Info("runEventsFormat done")
				return
			default:
				//log.Info("runEventsFormat default start")
				line, isPrefix, err := c.StdoutReader.ReadLine()
				if err != nil {
					if err != io.EOF {
						log.Error("c.StdoutReader.ReadLine err: %s", err.Error())
					}
				}
				if len(line) > 0 {
					var info ContainerInfo
					if err = json.Unmarshal(line, &info); err != nil {
						log.Error("json.Unmarshal err: %s", err.Error())
					} else {
						_, okC := c.ContainerMap[strings.ToLower(info.Name)]
						_, okS := c.StatusMap[strings.ToLower(info.Status)]

						if okC && okS {
							timeEvent := time.Unix(info.Time, 0)
							txt := fmt.Sprintf("服务：%s\n事件：%s\n时间：%s", info.Name, info.Status, timeEvent.Format("2006-01-02 15-04-05"))
							notify.SendLarkTextNotify(config.Cfg.Server.LarkNotifyKey, title, txt)
						} else if okC {
							log.Info("runEventsFormat:", info)
						}
					}
				} else {
					log.Info("runEventsFormat:", string(line), isPrefix)
				}
				//log.Info("runEventsFormat default end")
				time.Sleep(time.Second)
			}
		}
	}()
}
