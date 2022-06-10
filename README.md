# docker-events-monitor

### listen docker events
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
git clone https://github.com/scorpiotzh/docker-events-monitor.git
cd docker-event-monitor
make monitor
cp config/config.example.yaml config/config.yaml 
vim config/config.yaml 
./docker-event-monitor --config=./config/config.yaml
``` 