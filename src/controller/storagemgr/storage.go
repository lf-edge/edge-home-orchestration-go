package storagemgr

import (
	"controller/storagemgr/storagedriver"
	"errors"
	"github.com/edgexfoundry/device-sdk-go"
	"github.com/edgexfoundry/device-sdk-go/pkg/startup"
	"log"
	"os"
)



const (
	dataStorageService = "datastorage"
	edgeDir = "/var/edge-orchestration"
	dataStorageFilePath = edgeDir + "/datastorage/configuration.toml"
)

type Storage interface{
	StartStorage() error
}

type StorageImpl struct {}

var (
	storageIns *StorageImpl
)

func init(){
	storageIns = &StorageImpl{}
}
func GetInstance() Storage {
	return storageIns
}
func (StorageImpl) StartStorage() error {
	log.Printf("IN StartStorage")
	if _, err := os.Stat(dataStorageFilePath); err == nil {
		sd := storagedriver.StorageDriver{}
		go startup.Bootstrap(dataStorageService, device.Version, &sd)
		log.Printf("Could Properly Initialize the StorageMgr")
		return nil
	}
	log.Printf("Could Not Properly Initialize the StorageMgr")
	return errors.New("could not initiate storageManager")
}