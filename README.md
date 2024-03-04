# events-monitor

## listen docker events

```shell
docker events --filter 'type=container' --format '{"status":"{{.Status}}","name":"{{.Actor.Attributes.name}}","time":{{.Time}}}'
```

### config/config.yaml

```yaml
server:
  name: "cross-tzh" # name of server 
  lark_notify_key: "" # lark notify key
  containers: # listen containers
    - "cross_node_server"
    - "cross_node_mysql"
  status: # listen status
    - "start"
    - "stop"
```

### build & run

```shell
git clone https://events-monitor.git
cd docker-event-monitor
make monitor
cp config/config.example.yaml config/config.yaml 
vim config/config.yaml 
./docker-event-monitor --config=./config/config.yaml
``` 

```shell
#docker
cd /mnt/server/docker_events/docker-events-monitor
git pull
make monitor
mv /mnt/server/docker_events/monitor_server /mnt/server/docker_events/monitor_server.bak
mv /mnt/server/docker_events/docker-events-monitor/docker-events-monitor /mnt/server/docker_events/monitor_server
supervisorctl restart monitor_server
#sup
cd /mnt/server/docker_events/docker-events-monitor
git pull
make sup
mv /mnt/server/docker_events/supervisor_server /mnt/server/docker_events/supervisor_server.bak
mv /mnt/server/docker_events/docker-events-monitor/supervisor-events-listening /mnt/server/docker_events/supervisor_server
supervisorctl restart sup_listener
```

```shell
#docker
[program:monitor_server]
command = /mnt/server/docker_events/monitor_server --config=/mnt/server/docker_events/config/config.yaml

autostart=true                ; start at supervisord start (default: true)
autorestart=true
user=root                   ; setuid to this UNIX account to run the program
startsecs=2
startretries=3

redirect_stderr=true          ; redirect proc stderr to stdout (default false)
stdout_logfile=/mnt/server/docker_events/logs/out.log        ; stdout log path, NONE for none; default AUTO
stdout_logfile_maxbytes=100MB   ; max # logfile bytes b4 rotation (default 50MB)
stdout_logfile_backups=20     ; # of stdout logfile backups (default 10)
stdout_capture_maxbytes=100MB   ; number of bytes in 'capturemode' (default 0)
stdout_events_enabled=false   ; emit events on stdout writes (default false)
stderr_logfile=/mnt/server/docker_events/logs/err.log        ; stderr log path, NONE for none; default AUTO
stderr_logfile_maxbytes=100MB   ; max # logfile bytes b4 rotation (default 50MB)
stderr_logfile_backups=20     ; # of stderr logfile backups (default 10)
stderr_capture_maxbytes=100MB   ; number of bytes in 'capturemode' (default 0)
stderr_events_enabled=false   ; emit events on stderr writes (default false)
# sup
[eventlistener:sup_listener]
command=/mnt/server/docker_events/supervisor_server -key XXXX
events=PROCESS_STATE,TICK_5

stderr_logfile=/mnt/server/docker_events/logs_sup/err.log        ; stderr log path, NONE for none; default AUTO
stderr_logfile_maxbytes=100MB   ; max # logfile bytes b4 rotation (default 50MB)
stderr_logfile_backups=20     ; # of stderr logfile backups (default 10)
stderr_capture_maxbytes=100MB   ; number of bytes in 'capturemode' (default 0)
```

## listen supervisor events

install
```shell
go install events-monitor/cmd/supervisor-events-listener@latest
```

supervisor config

```shell
[eventlistener:listener_sup]
command=/mnt/server/docker-events-monitor/supervisor-events-listener -key dasdaf1afda-32sdaf1hogd3-jkliqvjjj; the key replace your own lark notify key
events=PROCESS_STATE,TICK_5

stderr_logfile=/mnt/server/docker-events-monitor/logs_sup/err.log        ; stderr log path, NONE for none; default AUTO
stderr_logfile_maxbytes=100MB   ; max # logfile bytes b4 rotation (default 50MB)
stderr_logfile_backups=20     ; # of stderr logfile backups (default 10)
stderr_capture_maxbytes=100MB   ; number of bytes in 'capturemode' (default 0)
```

## listen systemd service event

before install, need install libsystemd-dev, here we take ubuntu as an example
```shell
apt-get install -y libsystemd-dev
```

install
```shell
go install events-monitor/cmd/systemd-events-listener@latest
```

start
```shell
systemd-events-listener -key dasdaf1afda-32sdaf1hogd3-jkliqvjjj -services nginx,supervisor
```