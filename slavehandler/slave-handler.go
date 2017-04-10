package slavehandler

import (
	"bytes"
	"encoding/json"
	logger "github.com/Sirupsen/logrus"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

type Slave struct {
	Url    string `json:"slave_url"`
	Port   int    `json:"slave_port"`
	Name   string `json:"slave_name"`
	Action string `json:"slave_action"`
}

type SlavesConfiguration struct {
	Slaves map[string]*Slave `json:"slaves"`
}

var slavesConfiguration *SlavesConfiguration
var slavesConfigLock sync.Mutex
var slaveIp string
var slaveMacAddress string

func GetSlaveIPMacAddess() (error, string, string) {

	if slaveIp == "" {
		interfaces, err := net.Interfaces()
		if err != nil {
			logger.Error("no local internet interfaces found with error " + err.Error())
			return err, "", ""
		}
		for _, i := range interfaces {
			if addrs, err := i.Addrs(); err != nil {
				logger.Error("no local addresses internet interfaces found with error " + err.Error())
				return err, "", ""
			} else {
				for _, addr := range addrs {
					var ip net.IP

					switch v := addr.(type) {
					case *net.IPNet:

						ip = v.IP
					case *net.IPAddr:
						ip = v.IP
					}
					if ip.To4() != nil && !ip.IsLoopback() {
						slaveIp = ip.To4().String()
						if netInterface, err := net.InterfaceByName(i.Name); err != nil {
							logger.Error("Error while getting interfaceByName with error: " + err.Error())
							return err, slaveIp, slaveMacAddress
						} else {
							slaveMacAddress = netInterface.HardwareAddr.String()
							return nil, slaveIp, slaveMacAddress
						}
					}
				}
			}
		}
	}
	return nil, slaveIp, slaveMacAddress

}

func GetSlaves() *SlavesConfiguration {
	if slavesConfiguration == nil {
		slavesConfiguration = &SlavesConfiguration{Slaves: make(map[string]*Slave, 0)}
	}
	return slavesConfiguration
}

func AddSlave(slave *Slave) error {
	slavesConfig := GetSlaves()
	slavesConfig.Slaves[slave.Name] = slave
	return slavesConfig.saveConfiguration()
}

func NewSlave(url string, port int, name string, action string) *Slave {
	slave := &Slave{
		Url:    url,
		Port:   port,
		Name:   name,
		Action: action,
	}
	slavesConfig := GetSlaves()
	slavesConfig.Slaves[name] = slave
	slavesConfig.saveConfiguration()
	return slave
}

func (s *SlavesConfiguration) saveConfiguration() error {
	slavesConfigLock.Lock()
	defer slavesConfigLock.Unlock()
	f, err := os.Create("slaves_configuration.json")
	if err != nil {
		logger.Error("Error while saving slaves configuration with error " + err.Error())
		return err
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(s); err != nil {
		logger.Error("Error while saving slaves configuration with error " + err.Error())
		return err
	}
	return nil
}

func RegisterToMaster(masterUri string, localPort int, localAction string) {
	go func() {
		for {
			logger.Info("Attempt to register to " + masterUri)
			if err, ip, macAddress := GetSlaveIPMacAddess(); err != nil {
				logger.Error("not enable to get local ip and macaddress")
			} else {
				conf := &Slave{
					Url:    "http://" + ip,
					Port:   localPort,
					Name:   macAddress,
					Action: localAction,
				}

				body, _ := json.Marshal(conf)
				logger.Info("Body to send ", *conf)
				response, err := http.Post(masterUri, "application/json", bytes.NewBuffer(body))
				if err == nil {
					msg, _ := ioutil.ReadAll(response.Body)
					if response.StatusCode != 200 {
						logger.Error("Bad response from master " + masterUri + " with response " + string(msg))
					} else {
						errorMsg := string(msg)
						if errorMsg != "\"ok\"" {
							logger.Error("Bad response from master " + masterUri + " with response " + errorMsg)
						} else {
							logger.Info("Ok registered to " + masterUri + " with response " + errorMsg)
						}
					}
					response.Body.Close()
				} else {
					logger.Error("Error while registering to " + masterUri + " with error " + err.Error())
				}
			}

			time.Sleep(time.Second * 30)
		}
	}()

}
