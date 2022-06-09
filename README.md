# docker-events-monitor

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

```shell
./docker-event-monitor --config=./config/config.yaml
``` 