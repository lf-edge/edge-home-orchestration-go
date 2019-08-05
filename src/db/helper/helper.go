package helper

import (
	"strings"

	errormsg "common/errormsg"
	errors "common/errors"
	"db/bolt/common"
	configurationdb "db/bolt/configuration"
	networkdb "db/bolt/network"
	servicedb "db/bolt/service"
)

var (
	confQuery    configurationdb.Query
	netQuery     networkdb.Query
	serviceQuery servicedb.Query
)

func init() {
	netQuery = networkdb.Query{}
	confQuery = configurationdb.Query{}
	serviceQuery = servicedb.Query{}
}

type MultipleBucketQuery interface {
	GetDeviceInfoWithService(serviceName string, executionTypes []string) ([]ExecutionCandidate, error)
}

type ExecutionCandidate struct {
	Id       string
	ExecType string
	Endpoint []string
}

type multipleBucketQuery struct{}

var query multipleBucketQuery

func GetInstance() MultipleBucketQuery {
	return query
}

func (multipleBucketQuery) GetDeviceInfoWithService(serviceName string, executionTypes []string) ([]ExecutionCandidate, error) {
	ret := make([]ExecutionCandidate, 0)

	serviceItems, err := serviceQuery.GetList()
	if err != nil {
		return nil, err
	}

	for _, item := range serviceItems {
		for _, service := range item.Services {
			if strings.Compare(serviceName, service) == 0 {
				confItems, err := confQuery.Get(item.ID)
				if err != nil {
					continue
				}

				hasExecType := common.HasElem(executionTypes, confItems.ExecType)
				if hasExecType == false {
					continue
				}

				netItems, err := netQuery.Get(item.ID)
				if err != nil {
					continue
				}

				endpoints := make([]string, len(netItems.IPv4))
				copy(endpoints, netItems.IPv4)

				info := ExecutionCandidate{
					Id:       item.ID,
					ExecType: confItems.ExecType,
					Endpoint: endpoints,
				}

				ret = append(ret, info)
			}
		}
	}

	if len(ret) == 0 {
		err = errors.NotFound{Message: errormsg.ToString(errormsg.ErrorNoDeviceReturn)}
		return nil, err
	}

	return ret, nil
}
