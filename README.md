## Usage

### Controlling Units

```
curl -v http://127.0.0.1:8080/units/dnsmasq.service/start/replace
curl http://127.0.0.1:8080/units/dnsmasq.service/stop/replace
```

### Pulling images from a registry

```
curl localhost:8080/docker/registry/pull/philips/nova-agent
```
