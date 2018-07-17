package vmfloaty

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/user"

	"encoding/json"

	"gopkg.in/yaml.v2"
)

type FloatyConfig struct {
	URL   string
	User  string
	Token string
}

type PoolerClient struct {
	client http.Client
	config FloatyConfig
}

func NewPoolerClient(config FloatyConfig) PoolerClient {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := http.Client{Transport: tr}

	return PoolerClient{
		client: client,
		config: config,
	}
}

func Status(pooler PoolerClient) error {
	var data map[string]interface{}
	return get(&pooler.client, fmt.Sprintf("%s/status", pooler.config.URL), &data)
}

func List(pooler PoolerClient) ([]string, error) {
	var machines []string
	err := get(&pooler.client, fmt.Sprintf("%s/vm", pooler.config.URL), &machines)
	if err != nil {
		return machines, err
	}
	return machines, nil
}

func ListActive(pooler PoolerClient) []string {
	status := TokenStatus(pooler)
	return status.Detail.VMs.Running
}

type Host struct {
	Hostname string
	Domain   string
}

// Create is the same as retrive in the ruby version of the client
func Create(pooler PoolerClient, os string) (Host, error) {
	var data map[string]interface{}
	host := Host{}
	osString := os + "+"
	err := post(&pooler.client, fmt.Sprintf("%s/vm/%s", pooler.config.URL, osString), pooler.config.Token, &data)
	if err != nil {
		return host, err
	}

	if !processResponse(data, os, &host) {
		return host, fmt.Errorf("Error processing response")
	}

	if domain, ok := data["domain"]; ok {
		host.Domain = domain.(string)
	}

	return host, nil
}

func Delete(pooler PoolerClient, hostname string) error {
	return delete(&pooler.client, fmt.Sprintf("%s/vm/%s", pooler.config.URL, hostname), pooler.config.Token)
}

func processResponse(resp map[string]interface{}, key string, thing interface{}) bool {
	if value, ok := resp["ok"]; ok {
		if value == true {
			if item, ok := resp[key]; ok {
				data, err := json.Marshal(item)

				if err == nil {
					err := json.Unmarshal(data, &thing)
					if err == nil {
						return true
					}
				}
			}
		}
	}
	return false
}

func LoadConfig(filename string) (FloatyConfig, error) {
	c := FloatyConfig{}

	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	yamlFile, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", usr.HomeDir, filename))
	if err != nil {
		return c, err
	}
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		return c, err
	}

	return c, nil
}
