package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	cms "github.com/xumak-grid/init-containers/pkg/commons"
)

const (
	nexusUserEnv = "NEXUS_USER"
	nexusPassEnv = "NEXUS_PASS"
	nexusHostEnv = "NEXUS_HOST"
	// this file location contains configuration to make a initial setup to Nexus server
	nexusConfigFileEnv = "NEXUS_CONFIG_FILE"
	// nexusTimeout represents the maximun timeout to wait for the nexus
	nexusTimeout = "NEXUS_TIMEOUT"
)

// NexusConfig represents a valid configuration to create a resource in nexus,
// this is the object to POST in nexus server
type NexusConfig struct {
	Action string `json:"action"`
	Method string `json:"method"`
	// Data receives a slice of DataConfig but to make a POST request only one
	// object must be added.
	Data []DataConfig `json:"data"`
	Type string       `json:"type"`
	TID  int          `json:"tid"`
}

// Response represents the response of the request in the Nexus server
type Response struct {
	// Result is a result of the POST, when an error is present
	// the Suceess is equals to false
	Result *struct {
		Success bool `json:"success"`
	} `json:"result,omitempty"`
}

// DataConfig represents a configuration data for each POST in nexus
type DataConfig struct {
	Attributes          Attributes `json:"attributes"`
	Name                string     `json:"name"`
	Format              string     `json:"format"`
	Type                string     `json:"type"`
	URL                 string     `json:"url"`
	Online              bool       `json:"online"`
	AuthEnabled         bool       `json:"authEnabled,omitempty"`
	HTTPRequestSettings bool       `json:"httpRequestSettings,omitempty"`
	Recipe              string     `json:"recipe"`
}

// Attributes represents attributes from the data config.
type Attributes struct {
	Storage       Storage        `json:"storage,omitempty"`
	Group         *Group         `json:"group"`
	Maven         *Maven         `json:"maven,omitempty"`
	Proxy         *Proxy         `json:"proxy,omitempty"`
	HTTPClient    *HTTPClient    `json:"httpclient,omitempty"`
	NegativeCache *NegativeCache `json:"negativeCache,omitempty"`
}

// Storage represents storage options as part of attributes.
type Storage struct {
	BlobStoreName               string `json:"blobStoreName"`
	StrictContentTypeValidation bool   `json:"strictContentTypeValidation"`
	WritePolicy                 string `json:"writePolicy,omitempty"`
}

// Group represents a group of repositories.
type Group struct {
	MemberNames []string `json:"memberNames"`
}

// Maven represents a maven repository definition.
type Maven struct {
	VersionPolicy string `json:"versionPolicy"`
	LayoutPolicy  string `json:"layoutPolicy"`
}

// Proxy represents a proxy configuration for a repository.
type Proxy struct {
	RemoteURL      string `json:"remoteUrl"`
	ContentMaxAge  int    `json:"contentMaxAge"`
	MetadataMaxAge int    `json:"metadataMaxAge"`
}

// HTTPClient represents the http configuration.
type HTTPClient struct {
	Blocked        bool            `json:"blocked"`
	AutoBlock      bool            `json:"autoBlock"`
	Authentication *Authentication `json:"authentication"`
}

// NegativeCache contains cache specific settings.
type NegativeCache struct {
	Enabled    bool `json:"enabled"`
	TimeToLive int  `json:"timeToLive"`
}

// Authentication groups auth settings for a repository.
type Authentication struct {
	Type       string `json:"type"`
	UserName   string `json:"username"`
	Password   string `json:"password"`
	NtlmHost   string `json:"ntlmHost"`
	NtlmDomain string `json:"ntlmDomain"`
}

// nexusReady checks if nexus is ready to receive the POST commands
func nexusReady(user, pass, host string) bool {

	host = host + "/service/metrics/ping"
	req, err := http.NewRequest(http.MethodGet, host, nil)
	if err != nil {
		return false
	}

	req.SetBasicAuth(user, pass)
	client := cms.GetClient(5)
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true
	}

	return false
}

// nexusPost creates a new resource making a POST request to nexus
func nexusPost(user, pass, host string, obj NexusConfig) error {

	host = host + "/service/extdirect"
	jsonData, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, host, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.SetBasicAuth(user, pass)
	req.Header.Add("Content-Type", "application/json")
	client := cms.GetClient(5)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error creating resource code: %v message: %v", resp.StatusCode, body)
	}

	nexusResponse := Response{}
	err = json.Unmarshal(body, &nexusResponse)
	if err != nil {
		return err
	}
	if !nexusResponse.Result.Success {
		return fmt.Errorf("error: %v \nPOST: %v on data: %v", string(body), host, string(jsonData))
	}

	return nil
}

func main() {

	// environment variables
	configFile := cms.GetEnv(nexusConfigFileEnv, "examples/configFile.json")
	user := cms.GetEnv(nexusUserEnv, "admin")
	pass := cms.GetEnv(nexusPassEnv, "admin123")
	host := cms.GetEnv(nexusHostEnv, "http://localhost:8081")

	// read config file
	data := ArtifactoryConfig{}
	err := cms.DecodeFromFile(configFile, &data)
	if err != nil {
		log.Fatalf("reading configFile %v", err.Error())
	}

	log.Printf("check and wait for nexus on host: %v", host)

	timeout := time.After(1 * time.Minute)
	check := true
	for check {
		select {
		case <-time.After(3 * time.Second):
			if nexusReady(user, pass, host) {
				check = false
				break
			}
		case <-timeout:
			log.Fatalf("timeout reached, host: %v", host)
		}
		log.Println("host not ready, 3s")
	}

	log.Printf("installing (%d) hosted repositories", len(data.Hosteds))
	for _, h := range data.Hosteds {
		nexusConfig := nexusConfig(hostedDataConfig(h))
		err := nexusPost(user, pass, host, nexusConfig)
		if err != nil {
			log.Println(err.Error())
		}
		log.Printf("repository '%v' created\n", h.Name)
	}

	log.Printf("installing (%d) proxy repositories", len(data.Proxies))
	for _, p := range data.Proxies {
		nexusConfig := nexusConfig(proxyDataConfig(p))
		err := nexusPost(user, pass, host, nexusConfig)
		if err != nil {
			log.Println(err.Error())
		}
		log.Printf("repository '%v' created\n", p.Name)
	}

	log.Printf("installing (%d) group repositories", len(data.Groups))
	for _, g := range data.Groups {
		nexusConfig := nexusConfig(groupDataConfig(g))
		err := nexusPost(user, pass, host, nexusConfig)
		if err != nil {
			log.Println(err.Error())
		}
		log.Printf("repository '%v' created\n", g.Name)
	}

	log.Println("the job has finished successfully!")
}
