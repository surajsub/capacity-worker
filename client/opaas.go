package client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/spf13/viper"
)

type Patch struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value int    `json:"value"`
}

type OpaasData struct {
	Instances    []Instance    `json:"Instances"`
	Storage      []Storage     `json:"Storage"`
	Clusters     []Cluster     `json:"Clusters"`
	Clusterhosts []Clusterhost `json:"Clusterhosts"`
}

type OpaasApi struct {
	config *opaasConfig
}

type opaasConfig struct {
	baseURL string
	token   string
}

func NewOpaasApi() *OpaasApi {
	opaasConfig := getOpaasConfig()
	return &OpaasApi{
		config: opaasConfig,
	}
}

func getOpaasConfig() *opaasConfig {
	return &opaasConfig{
		baseURL: viper.GetString("OPAAS_BASE_URL"),
		token:   viper.GetString("OPAAS_APIKEY"),
	}
}

func (opaasApi *OpaasApi) get(endpoint string, output interface{}) error {
	data, httpErr := opaasApi.makeOpaasHTTPRequest("GET", endpoint, nil)
	if httpErr != nil {
		return httpErr
	}
	return json.Unmarshal(data, output)
}

func (opaasApi *OpaasApi) patch(model string, id string, patches []Patch) error {
	endpoint := fmt.Sprintf("%s/%s", model, id)
	patchBytes, marshalError := json.Marshal(patches)
	if marshalError != nil {
		return marshalError
	}
	patchReader := bytes.NewReader(patchBytes)
	_, httpErr := opaasApi.makeOpaasHTTPRequest("PATCH", endpoint, patchReader)
	return httpErr
}

func (opaasApi *OpaasApi) makeOpaasHTTPRequest(verb string, endpoint string, data io.Reader) ([]byte, error) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	request := opaasApi.createHTTPRequest(verb, endpoint, data)
	client := http.DefaultClient
	response, requestError := client.Do(request)
	if requestError != nil {
		return nil, requestError
	}
	if response.StatusCode != 200 {
		errMessage := fmt.Sprintf("Received non-200 status code: %s", response.Status)
		return nil, errors.New(errMessage)
	}
	defer response.Body.Close()
	return ioutil.ReadAll(response.Body)
}

func (opaasApi *OpaasApi) createHTTPRequest(verb string, endpoint string, data io.Reader) *http.Request {
	url := opaasApi.constructURL(endpoint)
	request, requestCreationErr := http.NewRequest(verb, url, data)
	if requestCreationErr != nil {
		panic(requestCreationErr)
	}
	opaasApi.addHeadersToRequest(verb, request)
	return request
}

func (opaasApi *OpaasApi) constructURL(endpoint string) string {
	baseURL := opaasApi.config.baseURL
	return fmt.Sprintf("%s/%s", baseURL, endpoint)
}

func (opaasApi *OpaasApi) addHeadersToRequest(verb string, request *http.Request) {
	bearerToken := opaasApi.config.token
	authHeader := fmt.Sprintf("Bearer %s", bearerToken)
	request.Header.Add("Authorization", authHeader)

	if verb == "PATCH" || verb == "POST" {
		request.Header.Set("Content-Type", "application/json")
	}

}
