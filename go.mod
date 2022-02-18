module github.com/lf-edge/edge-home-orchestration-go

go 1.16

require (
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/Microsoft/go-winio v0.4.16 // indirect
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/casbin/casbin v1.9.1
	github.com/containerd/containerd v1.4.12 // indirect
	github.com/docker/cli v0.0.0-20201024074417-fd3371eb7df1
	github.com/docker/distribution v2.8.0+incompatible // indirect
	github.com/docker/docker v17.12.0-ce-rc1.0.20201201034508-7d75c1d40d88+incompatible
	github.com/docker/go-connections v0.4.0
	github.com/docker/go-units v0.4.0
	github.com/eclipse/paho.mqtt.golang v1.3.5
	github.com/edgexfoundry/device-sdk-go v1.4.0
	github.com/edgexfoundry/go-mod-core-contracts v0.1.115
	github.com/fsnotify/fsnotify v1.4.9
	github.com/go-ole/go-ole v1.2.5 // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang-jwt/jwt/v4 v4.3.0
	github.com/golang/mock v1.4.4
	github.com/gomodule/redigo v1.8.3
	github.com/gorilla/mux v1.8.0
	github.com/grandcat/zeroconf v1.0.0
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/leemcloughlin/logfile v0.0.0-20201123203928-cff1c8a30a10
	github.com/mattn/go-shellwords v1.0.10 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.2 // indirect
	github.com/pelletier/go-toml v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/satori/go.uuid v1.2.0
	github.com/shirou/gopsutil v3.20.11+incompatible
	github.com/sirupsen/logrus v1.7.0
	github.com/smartystreets/goconvey v1.6.4 // indirect
	github.com/songgao/water v0.0.0-20200317203138-2b4b6d7c09d8
	github.com/spf13/cast v1.3.1
	github.com/spf13/pflag v1.0.3
	github.com/stretchr/testify v1.6.1
	github.com/vishvananda/netlink v1.1.0
	github.com/xeipuuv/gojsonpointer v0.0.0-20180127040702-4e3ac2762d5f // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v0.0.0-20160323030313-93e72a773fad // indirect
	go.etcd.io/bbolt v1.3.5
	golang.org/x/time v0.0.0-20210611083556-38a9dc6acbc6 // indirect
	google.golang.org/grpc v1.34.0 // indirect
	gopkg.in/ini.v1 v1.62.0
	gopkg.in/yaml.v2 v2.4.0
	gotest.tools v2.2.0+incompatible // indirect
	gotest.tools/v3 v3.0.3
)

replace github.com/grandcat/zeroconf v1.0.0 => ./third_party/zeroconf
