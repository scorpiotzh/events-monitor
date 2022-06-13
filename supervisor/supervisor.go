package supervisor

import (
	"bufio"
	"fmt"
	"github.com/dotbitHQ/docker-events-monitor/notify"
	"github.com/scorpiotzh/toolib"
	"net"
	"os"
	"strconv"
	"strings"
)

type EventsListener struct {
	Key    string
	stdin  *bufio.Reader
	stdout *bufio.Writer
	stderr *bufio.Writer
}

func (e *EventsListener) Run() {
	e.init()
	for {
		e.logErr(fmt.Errorf("e.parse() start"))
		e.ready()
		e.logErr(fmt.Errorf("e.parse() ready"))
		if err := e.parse(); err != nil {
			e.logErr(fmt.Errorf("e.parse() err: %s", err.Error()))
			e.parseFail()
		} else {
			e.logErr(fmt.Errorf("e.parse() ok"))
			e.parseOk()
		}
		//time.Sleep(time.Second)
	}
}

const RESP_OK = "RESULT 2\nOK"

func (e *EventsListener) Run2() {
	e.init()
	for {
		// 发送后等待接收event
		e.ready()
		// 接收header
		h, err := e.parseHeader()
		if err != nil {
			e.logErr(fmt.Errorf("e.parseHeader() err: %s", err.Error()))
		}
		e.logErr(fmt.Errorf("e.parseHeader():", toolib.JsonString(&h)))

		//line, _, _ := e.stdin.ReadLine()
		//_, _ = e.stderr.WriteString("read" + string(line))
		//_ = e.stderr.Flush()
		//
		//_, payloadSize := praseHeader(line)

		// 接收payload
		payload := make([]byte, h.Len)
		_, _ = e.stdin.Read(payload)
		_, _ = e.stderr.WriteString("read : " + string(payload))
		_ = e.stderr.Flush()

		e.parseOk()
	}
}

func praseHeader(data []byte) (header map[string]string,
	payloadSize int) {
	pairs := strings.Split(string(data), " ")
	header = make(map[string]string, len(pairs))

	for _, pair := range pairs {
		token := strings.Split(pair, ":")
		header[token[0]] = token[1]
	}

	payloadSize, _ = strconv.Atoi(header["len"])
	return header, payloadSize
}

func (e *EventsListener) parse() error {
	h, err := e.parseHeader()
	if err != nil {
		return fmt.Errorf("e.parseHeader err: %s", err.Error())
	}

	e.logErr(fmt.Errorf("e.parseHeader(): %s", toolib.JsonString(&h)))

	if h.Len == 0 {
		return nil
	}
	p, err := e.parsePayload(h.Len)
	if err != nil {
		return fmt.Errorf("e.parsePayload err: %s", err.Error())
	}

	e.logErr(fmt.Errorf("e.parsePayload(): %s", toolib.JsonString(&p)))

	e.sendLarkNotify(h, p)

	return nil
}

func (e *EventsListener) sendLarkNotify(h *Header, p *Payload) {
	title := "Supervisor 服务监控"
	text := fmt.Sprintf(`程序名称：%s
事件内容：%s
程序原状态：%s
服务器IP：%s
进程号：%d
`, p.ProcessName, h.EventName, p.FromState, p.Ip, p.Pid)

	switch h.EventName {
	case "PROCESS_STATE_STOPPED", "PROCESS_STATE_RUNNING":
		notify.SendLarkTextNotify(e.Key, title, text)
	default:
		notify.SendLarkTextNotify(e.Key, title, text)
		e.logErr(fmt.Errorf("sendLarkNotify: %s", h.EventName))
	}
}

func (e *EventsListener) init() {
	e.stdin = bufio.NewReader(os.Stdin)
	e.stdout = bufio.NewWriter(os.Stdout)
	e.stderr = bufio.NewWriter(os.Stderr)
}

func (e *EventsListener) ready() {
	_, _ = e.stdout.WriteString("READY\n")
	_ = e.stdout.Flush()
}

type Header struct {
	Ver        string `json:"ver"`
	Server     string `json:"server"`
	Serial     int    `json:"serial"`
	Pool       string `json:"pool"`
	PoolSerial int    `json:"pool_serial"`
	EventName  string `json:"event_name"` // 事件名称
	Len        int    `json:"len"`        // Payload长度
}

func (e *EventsListener) parseHeader() (*Header, error) {
	line, _, err := e.stdin.ReadLine()
	if err != nil {
		return nil, fmt.Errorf("e.Stdin.ReadLine err: %s", err.Error())
	}

	fields := e.parseFields(string(line))

	var h Header
	h.Ver = fields["ver"]
	h.Server = fields["server"]
	h.Serial, _ = strconv.Atoi(fields["serial"])
	h.Pool = fields["pool"]
	h.PoolSerial, _ = strconv.Atoi(fields["poolserial"])
	h.EventName = fields["eventname"]
	h.Len, _ = strconv.Atoi(fields["len"])

	return &h, nil
}

func (e *EventsListener) parseFields(data string) (fields map[string]string) {
	e.logErr(fmt.Errorf("parseFields: %s\n", data))
	fields = make(map[string]string)
	data = strings.TrimSpace(data)
	if data == "" {
		return
	}
	// ver:3.0 server:supervisor serial:5
	items := strings.Split(data, " ")
	if len(items) == 0 {
		return
	}
	for _, item := range items {
		values := strings.Split(item, ":")
		if len(values) < 2 {
			continue
		}
		key := strings.TrimSpace(values[0])
		value := strings.TrimSpace(values[1])
		fields[key] = value
	}

	return fields
}

type Payload struct {
	Ip          string `json:"ip"`
	ProcessName string `json:"process_name"` // 进程名称
	GroupName   string `json:"group_name"`   // 进程组名称
	FromState   string `json:"from_state"`
	Expected    int    `json:"expected"`
	Pid         int    `json:"pid"`
}

func (e *EventsListener) parsePayload(payloadLen int) (*Payload, error) {
	buf := make([]byte, payloadLen)
	_, err := e.stdin.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("e.Stdin.Read err: %s", err.Error())
	}

	fields := e.parseFields(string(buf))

	var p Payload

	hostname, _ := os.Hostname()
	p.Ip = fmt.Sprintf("%s(%s)", getLocalIp(), hostname)
	p.ProcessName = fields["processname"]
	p.GroupName = fields["groupname"]
	p.FromState = fields["from_state"]
	p.Expected, _ = strconv.Atoi(fields["expected"])
	p.Pid, _ = strconv.Atoi(fields["pid"])

	return &p, nil
}

func getLocalIp() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func (e *EventsListener) parseOk() {
	_, _ = e.stdout.WriteString("RESULT 2\nOK")
	_ = e.stdout.Flush()
}

func (e *EventsListener) parseFail() {
	_, _ = e.stdout.WriteString("RESULT 4\nFAIL")
	_ = e.stdout.Flush()
}

func (e *EventsListener) logErr(err error) {
	_, _ = e.stderr.WriteString(err.Error() + "\n")
	_ = e.stderr.Flush()
}
