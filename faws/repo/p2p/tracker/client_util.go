package tracker

import (
	"io"
	"net/http"
)

func (client *Client) put(name string, body io.Reader) (status int, reply io.ReadCloser, err error) {
	var request *http.Request
	request, err = http.NewRequest("PUT", client.base_url+name, body)
	if err != nil {
		return
	}

	var response *http.Response
	response, err = client.web.Do(request)
	if err != nil {
		return
	}
	status = response.StatusCode
	reply = response.Body
	return
}

func (client *Client) get(name string) (status int, content_type string, reply io.ReadCloser, err error) {
	var request *http.Request
	request, err = http.NewRequest("GET", client.base_url+name, nil)
	if err != nil {
		return
	}

	var response *http.Response
	response, err = client.web.Do(request)
	if err != nil {
		return
	}
	content_type = response.Header.Get("Content-Type")
	status = response.StatusCode
	reply = response.Body
	return
}
