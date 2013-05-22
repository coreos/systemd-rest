## Usage

### Controlling Units

```
curl -v http://127.0.0.1:8080/units/dnsmasq.service/start/replace
curl http://127.0.0.1:8080/units/dnsmasq.service/stop/replace
```

### Pulling images from a registry

```
curl localhost:8080/docker/registry/pull/busybox
curl -F "image=busybox" localhost:8080/docker/container/create/busybox
systemd-nspawn -b -D /var/lib/containers/busybox
```
