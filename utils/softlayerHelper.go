package utils

import (
	"fmt"
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"encoding/json"
)

const (
	url1 string = "https://api.softlayer.com/rest/v3/SoftLayer_Account/Hardware.json"

)

type SoftLayerHosts struct {
	FQDN string `json:"fullyQualifiedDomainName"`
	ID   int    `json:"id"`
}

func GetSLData() []SoftLayerHosts {
	SlData := []SoftLayerHosts{}
	httpErr1 := getSlJson(url1, &SlData)
	if httpErr1 != nil {
		fmt.Println("http error", httpErr1)
		return nil
	}
	SlData2 := []SoftLayerHosts{}
	httpErr2 := getSlJson(url1, &SlData2)
	if httpErr2 != nil {
		fmt.Println("http error", httpErr2)
		return nil
	}
	SlData = append(SlData, SlData2...)
	return SlData
}

func getSlJson(url string, output interface{}) error {
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true,}}
	client := &http.Client{Transport: tr}
	req, httpErr := http.NewRequest("GET", url, nil)
	if httpErr != nil {
		return httpErr
	}
	res, httpErr := client.Do(req)
	if httpErr != nil {
		return httpErr
	}

	defer res.Body.Close()

	body, httpErr := ioutil.ReadAll(res.Body)
	if httpErr != nil {
		return httpErr
	}
	return json.Unmarshal(body, output)
}
