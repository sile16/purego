
//Package purego Wrapper for the Pure Storage API
package purego

import (
	"fmt"
	"bytes"
	"net/http"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http/cookiejar"
)

//Client Create a new client
type Client struct {
	url string
	arrayHost string
	Username string
	Password string
	APIToken string

	apiVersion string
	httpClient *http.Client
	cookieJar *cookiejar.Jar
}

//NewClientUserPass New Client from User/Pass
func NewClientUserPass(arrayHost, username, password string) *Client {
	c := &Client {
		arrayHost: arrayHost,
		Username: username,
		Password: password,
	}
	c.init()
	return c
}

//NewClientAPIToken New Client from API Token
func NewClientAPIToken(arrayHost, apiToken string) *Client {
	c := &Client{
		arrayHost: arrayHost,
		APIToken: apiToken,
	}
	c.init()
	return c
}

func (c *Client) init() *Client {
	c.apiVersion = "1.12"
	c.url = "https://"+c.arrayHost+"/api/1.12/"
	c.cookieJar,_ = cookiejar.New(nil)
	c.httpClient = &http.Client{Jar: c.cookieJar}
	if c.APIToken == "" {
		c.StartSession()
	}
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	

	return c
}

//apiCall internal api call
func (c *Client) apiCall(method string,endPoint string, jsonData map[string]string ) (string, error) {
	url := c.url + endPoint
	fmt.Println(url)
	jsonValue, _ := json.Marshal(jsonData)
	req, err := http.NewRequest(method,url,bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	data := string(body)
	if err != nil {
		return "", err
	}
	if 200 != resp.StatusCode {
		return "", fmt.Errorf("%s", data)
	}
	return data, nil
}

//StartSession start session
func (c *Client) StartSession ()  {
	jsonData := map[string]string{"api_token": c.APIToken}
    
	data,err := c.apiCall("POST","auth/session",jsonData)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(data)
}

//GetArray get arrays
func (c *Client) GetArray() {
	fmt.Println(c.apiCall("GET","array",nil))
}