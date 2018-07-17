package vmfloaty

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func get(client *http.Client, url string, data interface{}) error {
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	handleResponse(resp, &data)

	return nil
}
func post(client *http.Client, url string, token string, data interface{}) error {

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-AUTH-TOKEN", token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)

	if err != nil {
		return err
	}

	handleResponse(resp, &data)

	return nil
}

func delete(client *http.Client, url string, token string) error {

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("X-AUTH-TOKEN", token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(resp.Status)
	}

	return nil

}

func handleResponse(response *http.Response, data interface{}) {

	if response.StatusCode == http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(response.Body)
		err := json.Unmarshal(bodyBytes, &data)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Fatal(response.StatusCode)
	}
}
