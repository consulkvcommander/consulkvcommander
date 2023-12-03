package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type ConsulKVClient struct {
	url string
}

func NewConsulKV(url string) ConsulKVClient {
	return ConsulKVClient{
		url: url,
	}
}

type ConsulKVResponseElement struct {
	Key   string `json:"Key,omitempty"`
	Value string `json:"Value,omitempty"`
}

type ConsulKVResponse []ConsulKVResponseElement

func (c ConsulKVClient) GetPath(path string) (ConsulKVResponse, error) {
	if len(path) == 0 {
		return ConsulKVResponse{}, nil
	}
	isDirectory := string(path[len(path)-1]) == "/"
	var apiEndpoint string
	if isDirectory {
		apiEndpoint = fmt.Sprintf("%s/v1/kv/%s?recurse=true", c.url, path)
	} else {
		apiEndpoint = fmt.Sprintf("%s/v1/kv/%s", c.url, path)
	}
	resp, statusCode, err := CallAPI(APIRequest{
		URL:         apiEndpoint,
		Method:      GET,
		ContentType: JSON,
	})
	if err != nil {
		return ConsulKVResponse{}, fmt.Errorf("error occurred while GET-ing from the ConsulKV client")
	}
	if statusCode != http.StatusOK && statusCode != http.StatusNotFound {
		return ConsulKVResponse{}, fmt.Errorf("failed to GET from the ConsulKV Client (status code '%d'): %s", statusCode, string(resp))
	}
	var consulKvResponse []ConsulKVResponseElement
	if resp == nil || string(resp) == "" {
		return consulKvResponse, nil
	}
	if err := json.Unmarshal(resp, &consulKvResponse); err != nil {
		return ConsulKVResponse{}, fmt.Errorf("failed to parse the response from Consul KV for the path '%s': %w", path, err)
	}
	return consulKvResponse, nil
}

func (c ConsulKVClient) DeletePath(path string) error {
	if len(path) == 0 {
		return nil
	}
	isDirectory := string(path[len(path)-1]) == "/"
	var apiEndpoint string
	if isDirectory {
		apiEndpoint = fmt.Sprintf("%s/v1/kv/%s?recurse=true", c.url, path)
	} else {
		apiEndpoint = fmt.Sprintf("%s/v1/kv/%s", c.url, path)
	}
	resp, statusCode, err := CallAPI(APIRequest{
		URL:         apiEndpoint,
		Method:      DELETE,
		ContentType: JSON,
	})
	if err != nil {
		return fmt.Errorf("error occurred while DELETE-ing from the ConsulKV client")
	}
	if statusCode != http.StatusOK && statusCode != http.StatusNotFound {
		return fmt.Errorf("failed to DELETE from the ConsulKV Client (status code '%d'): %s", statusCode, string(resp))
	}
	return nil
}

func (c ConsulKVClient) UpdatePath(path string, newValue string) error {
	if len(path) == 0 {
		return nil
	}
	isDirectory := string(path[len(path)-1]) == "/"
	var apiEndpoint string
	if isDirectory {
		apiEndpoint = fmt.Sprintf("%s/v1/kv/%s?recurse=true", c.url, path)
	} else {
		apiEndpoint = fmt.Sprintf("%s/v1/kv/%s", c.url, path)
	}
	resp, statusCode, err := CallAPI(APIRequest{
		URL:    apiEndpoint,
		Method: PUT,
		Body:   bytes.NewReader([]byte(newValue)),
	})
	if err != nil {
		return fmt.Errorf("error occurred while PUT-ing from the ConsulKV client")
	}
	if statusCode != http.StatusOK {
		return fmt.Errorf("failed to PUT from the ConsulKV Client (status code '%d'): %s", statusCode, string(resp))
	}
	return nil
}
