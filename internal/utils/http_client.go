package utils

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

type APIMethod string

const (
	GET    APIMethod = "GET"
	POST   APIMethod = "POST"
	PUT    APIMethod = "PUT"
	DELETE APIMethod = "DELETE"
	PATCH  APIMethod = "PATCH"
)

type APIContentType string

const (
	JSON APIContentType = "application/json"
	XML  APIContentType = "application/xml"
	FORM APIContentType = "application/x-www-form-urlencoded"
)

type APIRequest struct {
	URL         string
	Method      APIMethod
	ContentType APIContentType
	Body        io.Reader
}

func CallAPI(request APIRequest) ([]byte, int, error) {
	var response *http.Response
	switch request.Method {
	case GET:
		resp, err := http.Get(request.URL)
		if err != nil {
			return []byte{}, http.StatusInternalServerError, fmt.Errorf("error occurred while calling %s: %w", request.URL, err)
		}
		response = resp
	case POST, PUT, DELETE, PATCH:
		req, err := http.NewRequest(string(request.Method), request.URL, request.Body)
		if err != nil {
			return []byte{}, http.StatusInternalServerError, fmt.Errorf("error occurred while calling %s: %w", request.URL, err)
		}
		req.Header.Set("Content-Type", string(request.ContentType))
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return []byte{}, http.StatusInternalServerError, fmt.Errorf("error occurred while calling %s: %w", request.URL, err)
		}
		response = resp
	default:
		return []byte{}, http.StatusInternalServerError, fmt.Errorf("error occurred while calling %s: wrong request type found", request.URL)
	}

	defer response.Body.Close()
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return []byte{}, http.StatusInternalServerError, fmt.Errorf("error occurred while parsing the body: %w", err)
	}

	return responseBody, response.StatusCode, nil
}
