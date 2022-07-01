module github.com/lf-edge/edge-home-orchestration-go

go 1.16

require (
	github.com/Microsoft/go-winio v0.5.2 // indirect
	github.com/casbin/casbin v1.9.1
	github.com/containerd/containerd v1.5.13 // indirect
	github.com/docker/cli v20.10.17+incompatible
	github.com/docker/distribution v2.8.0+incompatible // indirect
	github.com/docker/docker v17.12.0-ce-rc1.0.20201201034508-7d75c1d40d88+incompatible
	github.com/docker/go-connections v0.4.0
	github.com/docker/go-units v0.4.0
	github.com/eclipse/paho.mqtt.golang v1.3.5
	github.com/edgexfoundry/device-sdk-go v1.4.0
	github.com/edgexfoundry/go-mod-core-contracts v0.1.115
	github.com/fsnotify/fsnotify v1.5.4
	github.com/golang-jwt/jwt/v4 v4.4.1
	github.com/golang/mock v1.4.4
	github.com/gomodule/redigo v1.8.8
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/grandcat/zeroconf v1.0.0
	github.com/imdario/mergo v0.3.13 // indirect
	github.com/leemcloughlin/logfile v0.0.0-20201123203928-cff1c8a30a10
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/pelletier/go-toml v1.9.5
	github.com/pkg/errors v0.9.1
	github.com/satori/go.uuid v1.2.0
	github.com/shirou/gopsutil v3.21.11+incompatible
	github.com/sirupsen/logrus v1.8.1
	github.com/songgao/water v0.0.0-20200317203138-2b4b6d7c09d8
	github.com/spf13/cast v1.4.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.8.0
	github.com/vishvananda/netlink v1.1.1-0.20210330154013-f5de75959ad5
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	go.etcd.io/bbolt v1.3.6
	golang.org/x/time v0.0.0-20220609170525-579cf78fd858 // indirect
	google.golang.org/grpc v1.47.0 // indirect
	gopkg.in/ini.v1 v1.66.6
	gopkg.in/yaml.v3 v3.0.1
	gotest.tools/v3 v3.2.0
)

replace github.com/grandcat/zeroconf v1.0.0 => ./third_party/zeroconf
