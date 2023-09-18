package systemd

import (
	"fmt"
	"github.com/coreos/go-systemd/v22/sdjournal"
	"github.com/scorpiotzh/docker-events-monitor/notify"
	"github.com/scorpiotzh/docker-events-monitor/tool"
	"github.com/scorpiotzh/mylog"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var log = tool.GetLog("systemd", mylog.LevelDebug)

type ServiceStatus string

const (
	FieldUnit      = "UNIT"
	FieldJobType   = "JOB_TYPE"
	FieldJobResult = "JOB_RESULT"
	FieldMessage   = "MESSAGE"

	JobResultDone   = "done"
	JobResultFailed = "failed"
	JobTypeStart    = "start"
	JobTypeStop     = "stop"
	JobTypeRestart  = "restart"
	JobTypeReload   = "reload"

	ServiceSuffix             = ".service"
	MaxReserveJournalEntryNum = 10

	ServiceStatusStarting     = "STARTING"
	ServiceStatusStarted      = "STARTED"
	ServiceStatusStartFailed  = "START_FAILED"
	ServiceStatusStopping     = "STOPPING"
	ServiceStatusStopped      = "STOPPED"
	ServiceStatusStopFailed   = "STOP_FAILED"
	ServiceStatusReloading    = "RELOADING"
	ServiceStatusReloaded     = "RELOADED"
	ServiceStatusReloadFailed = "RELOAD_FAILED"
)

type EventListener struct {
	Services []string
	LarkKey  string

	monitorServices map[string]*monitorServiceEntity
}

type monitorServiceEntity struct {
	status       ServiceStatus
	latestEntity []*sdjournal.JournalEntry
}

type Payload struct {
	Ip          string `json:"ip"`
	ProcessName string `json:"process_name"` // 进程名称
	FromState   string `json:"from_state"`
	EventName   string `json:"event_name"`
	Pid         int    `json:"pid"`
	Message     string `json:"message"`
}

func (e *EventListener) Run() error {
	e.monitorServices = make(map[string]*monitorServiceEntity)

	j, err := sdjournal.NewJournal()
	if err != nil {
		return err
	}
	if j == nil {
		log.Fatal("Got a nil journal")
	}
	defer j.Close()

	for _, v := range e.Services {
		svrName := strings.TrimSpace(v)
		if !strings.HasSuffix(svrName, ServiceSuffix) {
			svrName += ServiceSuffix
		}
		m := sdjournal.Match{Field: FieldUnit, Value: svrName}
		if err = j.AddMatch(m.String()); err != nil {
			return err
		}
	}

	if err := j.SeekTail(); err != nil {
		return err
	}
	if _, err := j.Next(); err != nil {
		return err
	}

	for {
		n, err := j.Next()
		if err != nil {
			log.Error(err)
			time.Sleep(time.Second)
			continue
		}
		if n == 0 {
			time.Sleep(time.Second)
			continue
		}
		entity, err := j.GetEntry()
		if err != nil {
			log.Error(err)
			time.Sleep(time.Second)
			continue
		}
		svrName := entity.Fields[FieldUnit]
		monitorSvr, ok := e.monitorServices[svrName]
		if !ok {
			monitorSvr = &monitorServiceEntity{
				latestEntity: make([]*sdjournal.JournalEntry, 0),
			}
			e.monitorServices[svrName] = monitorSvr
		}
		monitorSvr.latestEntity = append(monitorSvr.latestEntity, entity)
		if len(monitorSvr.latestEntity) > MaxReserveJournalEntryNum {
			monitorSvr.latestEntity = monitorSvr.latestEntity[len(monitorSvr.latestEntity)-MaxReserveJournalEntryNum:]
		}

		jobType := entity.Fields[FieldJobType]
		jobResult := entity.Fields[FieldJobResult]

		cmd := exec.Command("systemctl", "show", "--property", "MainPID", "--value", svrName)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return err
		}
		pid, err := strconv.Atoi(strings.TrimSpace(string(out)))
		if err != nil {
			return err
		}

		payload := &Payload{

			Ip:          fmt.Sprintf("%s(%s)", tool.GetLocalIp(), entity.Fields[sdjournal.SD_JOURNAL_FIELD_HOSTNAME]),
			ProcessName: svrName,
			Pid:         pid,
			EventName:   jobType,
		}

		switch jobType {
		case JobTypeStart:
			switch jobResult {
			case "":
				payload.FromState = ServiceStatusStarting
			case JobResultDone:
				payload.FromState = ServiceStatusStarted
			case JobResultFailed:
				payload.FromState = ServiceStatusStartFailed
				payload.Message = entity.Fields[FieldMessage]
			}
		case JobTypeStop:
			switch jobResult {
			case "":
				payload.FromState = ServiceStatusStopping
			case JobResultDone:
				payload.FromState = ServiceStatusStopped
			case JobResultFailed:
				payload.FromState = ServiceStatusStopFailed
				payload.Message = entity.Fields[FieldMessage]
			}
		case JobTypeRestart:
			switch jobResult {
			case JobResultDone:
				payload.FromState = ServiceStatusStopped
			case JobResultFailed:
				payload.FromState = ServiceStatusStopFailed
				payload.Message = entity.Fields[FieldMessage]
			}
		case JobTypeReload:
			switch jobResult {
			case "":
				payload.FromState = ServiceStatusReloading
			case JobResultDone:
				payload.FromState = ServiceStatusReloaded
			case JobResultFailed:
				payload.FromState = ServiceStatusReloadFailed
				payload.Message = entity.Fields[FieldMessage]
			}
		default:
			log.Warnf("job_type: %s no support", jobType)
			continue
		}
		e.sendNotify(payload)
	}
}

func (e *EventListener) sendNotify(p *Payload) {
	title := "Systemd 服务监控"
	text := fmt.Sprintf("程序名称：%s\n事件内容：%s\n程序原状态：%s\n服务器IP：%s\n进程号：%d",
		p.ProcessName, p.EventName, p.FromState, p.Ip, p.Pid)
	if p.Message != "" {
		text += fmt.Sprintf("\n报错内容：%s", p.Message)
	}
	notify.SendLarkTextNotify(e.LarkKey, title, text)
}
