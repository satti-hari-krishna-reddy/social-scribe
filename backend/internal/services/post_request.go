package services

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

func MakePostRequest(url string, body []byte, headers map[string]string) ([]byte, error) {
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %v", err)
	}

	for key, value := range headers {
		request.Header.Set(key, value)
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(response.Body)
		return nil, fmt.Errorf("GraphQL query failed with status code %d: %s", response.StatusCode, string(body))
	}

	return ioutil.ReadAll(response.Body)
}
