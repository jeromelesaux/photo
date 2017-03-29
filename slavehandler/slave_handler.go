package slavehandler

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"photo/logger"
	"strconv"
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
		logger.Log("Error while saving slaves configuration with error " + err.Error())
		return err
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(s); err != nil {
		logger.Log("Error while saving slaves configuration with error " + err.Error())
		return err
	}
	return nil
}

func RegisterToMaster(masterUri string, localPort int, localAction string) {
	go func() {
		processid := strconv.Itoa(os.Getpid())
		for {
			logger.Log("Attempt to register to " + masterUri)
			interfaces, err := net.Interfaces()
			if err != nil {
				logger.Log("no local internet interfaces found with error " + err.Error())
				return
			}
			for _, i := range interfaces {
				addrs, err := i.Addrs()
				if err != nil {
					logger.Log("no local addresses internet interfaces found with error " + err.Error())
					return
				}
				for _, addr := range addrs {
					var ip net.IP

					switch v := addr.(type) {
					case *net.IPNet:
						ip = v.IP
					case *net.IPAddr:
						ip = v.IP
					}
					if ip.To4() != nil {
						conf := &Slave{
							Url:    "http://" + ip.To4().String(),
							Port:   localPort,
							Name:   "photoexif-" + processid,
							Action: localAction,
						}

						body, _ := json.Marshal(conf)
						logger.LogLn("Body to send ", *conf)
						response, err := http.Post(masterUri, "application/json", bytes.NewBuffer(body))
						if err == nil {
							msg, _ := ioutil.ReadAll(response.Body)
							if response.StatusCode != 200 {
								logger.Log("Bad response from master " + masterUri + " with response " + string(msg))
							} else {
								if string(msg) != "ok" {
									logger.Log("Bad response from master " + masterUri + " with response " + string(msg))
								} else {
									logger.Log("Ok registered to " + masterUri + " with response " + string(msg))
								}
							}
							response.Body.Close()
						} else {
							logger.Log("Error while registering to " + masterUri + " with error " + err.Error())
						}

					}
				}

			}
			time.Sleep(time.Second * 30)
		}
	}()

}
