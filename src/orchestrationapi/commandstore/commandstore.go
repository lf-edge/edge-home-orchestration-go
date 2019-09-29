package commandstore

import (
	"errors"
	"sync"

	"common/types/configuremgrtypes"
)

type CommandList interface {
	GetServiceFileName(serviceName string) (string, error)
	StoreServiceInfo(serviceInfo configuremgrtypes.ServiceInfo)
}

type commandListImpl struct {
	serviceInfos map[string]string
	mutex        *sync.Mutex
}

var commandList commandListImpl

func GetInstance() *CommandList {
	return &commandList
}

func (c commandListImpl) GetServiceFileName(serviceName string) (string, error) {
	val, ok := c.serviceInfos[serviceName]
	if !ok {
		return "", errors.New("not found registered service")
	}

	return val, nil
}

func (c *commandListImpl) StoreServiceInfo(serviceInfo configuremgrtypes.ServiceInfo) {
	mutex.Lock()
	c.serviceInfos[serviceInfo.ServiceName] = serviceInfo.ExecutableFileName
	mutex.Unlock()
}
