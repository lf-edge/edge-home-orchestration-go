package helper

import (
	"strings"

	errormsg "github.com/lf-edge/edge-home-orchestration-go/src/common/errormsg"
	errors "github.com/lf-edge/edge-home-orchestration-go/src/common/errors"
	"github.com/lf-edge/edge-home-orchestration-go/src/db/bolt/common"
	configurationdb "github.com/lf-edge/edge-home-orchestration-go/src/db/bolt/configuration"
	networkdb "github.com/lf-edge/edge-home-orchestration-go/src/db/bolt/network"
	servicedb "github.com/lf-edge/edge-home-orchestration-go/src/db/bolt/service"
)

var (
	confQuery    configurationdb.DBInterface
	netQuery     networkdb.DBInterface
	serviceQuery servicedb.DBInterface
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

	confItems, err := confQuery.GetList()
	if err != nil {
		return nil, err
	}

	for _, confItem := range confItems {
		hasExecType := common.HasElem(executionTypes, confItem.ExecType)
		if hasExecType == false {
			continue
		}

		if confItem.ExecType == "container" {
			endpoints, err := getEndpoints(confItem.ID)
			if err != nil {
				continue
			}

			info := ExecutionCandidate{
				Id:       confItem.ID,
				ExecType: confItem.ExecType,
				Endpoint: endpoints,
			}

			ret = append(ret, info)
			continue
		}

		serviceItem, err := serviceQuery.Get(confItem.ID)
		if err != nil {
			return nil, err
		}

		for _, service := range serviceItem.Services {
			if strings.Compare(serviceName, service) == 0 {
				endpoints, err := getEndpoints(serviceItem.ID)
				if err != nil {
					continue
				}

				info := ExecutionCandidate{
					Id:       serviceItem.ID,
					ExecType: confItem.ExecType,
					Endpoint: endpoints,
				}

				ret = append(ret, info)
				break
			}
		}
	}

	if len(ret) == 0 {
		err = errors.NotFound{Message: errormsg.ToString(errormsg.ErrorNoDeviceReturn)}
		return nil, err
	}

	return ret, nil
}

func getEndpoints(id string) ([]string, error) {
	netItems, err := netQuery.Get(id)
	if err != nil {
		return nil, err
	}

	endpoints := make([]string, len(netItems.IPv4))
	copy(endpoints, netItems.IPv4)

	return endpoints, nil
}
