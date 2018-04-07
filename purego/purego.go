//Package purego Wrapper for the Pure Storage API
package purego

import (
	"time"

	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
)

//Client Create a new client
type Client struct {
	url       string
	arrayHost string
	Username  string
	Password  string
	APIToken  string

	authStatus int
	retryCount int
	retryMax   int

	apiVersion     string
	httpClient     *http.Client
	cookieJar      *cookiejar.Jar
	sessionStarted bool
	LogLevel       int //0 - error, 1 - Warning, 2 - Info, 3 - Debug

	httpShortTimeout time.Duration //used for authentication and session starting
	httpLongTimeout  time.Duration //used for all later API Calls.
}

//NewClientUserPassAPI new client
func NewClientUserPassAPI(arrayHost, username, password, apiToken string) *Client {
	c := &Client{
		arrayHost: arrayHost,
		Username:  username,
		Password:  password,
		APIToken:  apiToken,
	}
	c.init()
	return c
}

//NewClientUserPass New Client from User/Pass
func NewClientUserPass(arrayHost, username, password string) *Client {
	return NewClientUserPassAPI(arrayHost, username, password, "")
}

//NewClientAPIToken New Client from API Token
func NewClientAPIToken(arrayHost, apiToken string) *Client {
	return NewClientUserPassAPI(arrayHost, "", "", apiToken)
}

//NewClient New Client from API Token
func NewClient(arrayHost string) *Client {
	return NewClientUserPass(arrayHost, "pureuser", "pureuser")
}

//log
func (c *Client) log(level int, msg string) {
	if level <= c.LogLevel {
		fmt.Println(msg)
	}
}

func (c *Client) init() {
	c.apiVersion = "1.12"
	c.url = "https://" + c.arrayHost + "/api/" + c.apiVersion + "/"
	c.cookieJar, _ = cookiejar.New(nil)
	c.httpClient = &http.Client{Jar: c.cookieJar}
	c.sessionStarted = false
	c.LogLevel = 3
	c.httpShortTimeout = 3 * time.Second
	c.httpLongTimeout = 20 * time.Second

	//initially set timeout small so that network, auth happen quick or fail quick
	//After successful session we will raise timeout.
	c.httpClient.Timeout = c.httpShortTimeout

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
}

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

//apiCall internal api call, that unmarshalls JSON into specific type.
func (c *Client) apiCall(method string, endPoint string) error {
	return c.apiCallJSON(method,endPoint,nil,nil)
}


//apiCallJSON internal api call, that unmarshalls JSON into specific type.
func (c *Client) apiCallJSON(method string, endPoint string, reqData []byte, resInt interface{}) error {
	c.log(3, "API Call: "+endPoint)

	if !c.sessionStarted {
		//TODO: turn this back on afterd one testing the error handling below.
		err := c.StartSession()
		if err != nil {
			return err
		}
	}

	resp, err := c.doHTTPRequest(method, endPoint, reqData, resInt)
	if err != nil {
		return err
	}

	for retry := 0; retry < 2; retry++ {

		switch resp.StatusCode {
		case 401:
			//UNAUTHORIZED
			//This means we need to start our session.
			c.log(3, "Unauthrorized, restarting session and retrying")
			c.sessionStarted = false
			err = c.StartSession()
			if err != nil {
				return err
			}
			//then retry
			resp, err = c.doHTTPRequest(method, endPoint, reqData, resInt)
			break

		case 200:
			//Everything Okay Return
			return nil

		default:
			//all other error conditions
			return fmt.Errorf("Error: %d  Response: %s", resp.StatusCode, resp.Body)
		}

	}

	//Do one last check to see if there was an error
	if err != nil || resp.StatusCode != 200 {
		return fmt.Errorf("Error: Max retry hit last error: %s", err.Error())
	}

	//Everything looks okay return nil.
	return nil

}

//PureAuthSessionV1_12 response
type PureAuthSessionV1_12 struct {
	Username string      `json:"username"`
	Msg      interface{} `json:"msg"`
}

// PureAPITokenV1_12 response
type PureAPITokenV1_12 struct {
	APIToken string      `json:"api_token"`
	Msg      interface{} `json:"msg"`
}

//GetAPIToken get api token using username & password
func (c *Client) GetAPIToken() error {
	c.log(3, "GetAPIToken()")
	if c.Username == "" {
		return fmt.Errorf("GetAPIToken: No Username available to get API Token")
	}

	data := []byte(`{"username":"` + c.Username + `", "password":"` + c.Password + `"}`)

	result := PureAPITokenV1_12{}

	resp, err := c.doHTTPRequest("POST", "auth/apitoken", data, &result)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 || result.APIToken == "" {
		//error no matching user name found
		return fmt.Errorf("Error no API token returned when trying to get new API token: %s", resp.Body)

	}

	//success!
	c.APIToken = result.APIToken
	c.log(3, "Received API Token:"+result.APIToken)
	return nil

}

//StartSession start session
func (c *Client) StartSession() error {
	c.log(3, "StartSession()")
	//Some large API calls make take some time.
	c.httpClient.Timeout = c.httpShortTimeout

	if c.APIToken == "" {
		//we don't have an API token
		err := c.GetAPIToken()
		if err != nil {
			return err
		}
	}

	data := []byte(`{"api_token": "` + c.APIToken + `"}`)
	result := PureAuthSessionV1_12{}

	resp, err := c.doHTTPRequest("POST", "auth/session", data, &result)

	if err != nil {
		return err
	}

	if resp.StatusCode == 400 {
		//INVALID API Token , If we have user/pass try and fall back to that.
		retryErr := c.GetAPIToken()
		if retryErr != nil {
			return fmt.Errorf("Start Session, Bad Request: %s, retry, error %s", spew.Sdump(result.Msg), retryErr.Error())
		}

		//Retry now that we have a valid API Token
		data := []byte(`{"api_token": "` + c.APIToken + `"}`)
		resp, err = c.doHTTPRequest("POST", "auth/session", data, &result)

		if err != nil {
			return err
		}

	}

	if result.Username == "" {
		//error no matching user name found
		return fmt.Errorf("Error no username returned when starting session" + resp.Body)
	}

	c.log(3, "Started session as user: "+result.Username)
	c.sessionStarted = true

	//Some large API calls make take some time.
	c.httpClient.Timeout = c.httpLongTimeout
	return err
}

//PureArrayV1_12 pure array details
type PureArrayV1_12 struct {
	Version   string `json:"version"`
	Revision  string `json:"revision"`
	ArrayName string `json:"array_name"`
	ID        string `json:"id"`
}

//GetArray get array
func (c *Client) GetArray() PureArrayV1_12 {
	c.log(3, "Get Array")
	result := PureArrayV1_12{}
	err := c.apiCallJSON("GET", "array", nil, &result)
	if err != nil {
		c.log(0, err.Error())
	}
	//c.log(4,string(result))
	return result
}

//PureVolumesV1_12 d
type PureVolumesV1_12 []struct {
	Total             int64     `json:"total,omitempty"`
	Name              string    `json:"name"`
	System            int64     `json:"system,omitempty"`
	Snapshots         int64     `json:"snapshots,omitempty"`
	Volumes           int64     `json:"volumes,omitempty"`
	DataReduction     float64   `json:"data_reduction,omitempty"`
	Size              int64     `json:"size,omitempty"`
	SharedSpace       int64     `json:"shared_space,omitempty"`
	ThinProvisioning  float64   `json:"thin_provisioning,omitempty"`
	TotalReduction    float64   `json:"total_reduction,omitempty"`
	WritesPerSec      int       `json:"writes_per_sec,omitempty"`
	UsecPerWriteOp    int       `json:"usec_per_write_op,omitempty"`
	OutputPerSec      int       `json:"output_per_sec,omitempty"`
	SanUsecPerReadOp  int       `json:"san_usec_per_read_op,omitempty"`
	ReadsPerSec       int       `json:"reads_per_sec,omitempty"`
	InputPerSec       int       `json:"input_per_sec,omitempty"`
	Time              time.Time `json:"time,omitempty"`
	SanUsecPerWriteOp int       `json:"san_usec_per_write_op,omitempty"`
	UsecPerReadOp     int       `json:"usec_per_read_op,omitempty"`
	Source            string    `json:"source,omitempty"`
	Serial            string    `json:"serial,omitempty"`
	Created           time.Time `json:"created,omitempty"`
}

//GetVolumes tesitn
func (c *Client) GetVolumes() PureVolumesV1_12 {
	c.log(3, "GetVolumes)")
	result := PureVolumesV1_12{}
	err := c.apiCallJSON("GET", "volume", nil, &result)
	if err != nil {
		c.log(0, err.Error())
	}
	//c.log(4,string(result))
	return result
}
