package purego

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

//PureHTTPResponse returning info from http response, just provide what we need.
type PureHTTPResponse struct {
	Request    string
	Body       string
	StatusCode int
}

//doHTTPRequest internal api call
func (c *Client) doHTTPRequest(method string, endPoint string, reqData []byte, resInt interface{}) (*PureHTTPResponse, error) {
	//append endpoint to base url
	c.log(3, "doHTTP Request: "+endPoint)
	url := c.url + endPoint
	//fmt.Println(url)

	//setup request
	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqData))
	req.Header.Set("Content-Type", "application/json")

	//Actually send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	//read response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	pureRes := PureHTTPResponse{Body: string(body),
		StatusCode: resp.StatusCode}

	//Unmarshal the response based on the passed interface
	if resInt != nil {
		jsonError := json.Unmarshal(body, &resInt)
		if jsonError != nil {
			return &pureRes, jsonError
		}

	}
	return &pureRes, nil
}