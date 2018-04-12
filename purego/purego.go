//Package purego Wrapper for the Pure Storage API
//This is unofficial and is actually the first program I've ever written in Golang :)
//Matt Robertson mrobertson@purestorage.com

package purego

import (
	"time"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/cookiejar"
	"sync"
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
	secure 	bool //will check TLS certificate when true.

	lastSessionUse time.Time
	sessionLock sync.Mutex
	maxAPICalls chan struct{}
}

func (c *Client) init() {
	c.apiVersion = "1.12"
	c.url = "https://" + c.arrayHost + "/api/" + c.apiVersion + "/"
	
	c.sessionStarted = false
	c.LogLevel = 1
	c.maxAPICalls = make(chan struct{},10)  //This sets the maximum number of concurrent API calls.


	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
		  Timeout: 2 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 2 * time.Second,
	}

	//Allow insecure certificates
	netTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: !c.secure}

	c.cookieJar, _ = cookiejar.New(nil)

	c.httpClient = &http.Client{
		Timeout: time.Second * 30,
		Transport: netTransport,
		Jar: c.cookieJar,
	}
}

//newClientUserPassAPISecure new client
func newClientUserPassAPISecure(arrayHost, username, password, apiToken string, secure bool) *Client {
	c := &Client{
		arrayHost: arrayHost,
		Username:  username,
		Password:  password,
		APIToken:  apiToken,
		secure: secure,
	}
	c.init()
	return c
}

//NewClientUserPassAPIInsecure New Client from User/Pass & API & bypass certificate check
func NewClientUserPassAPIInsecure(arrayHost, username, password, apiToken string) *Client {
	return newClientUserPassAPISecure(arrayHost, username, password, apiToken,false)
}

//NewClientUserPassAPI New Client from User/Pass  & API
func NewClientUserPassAPI(arrayHost, username, password , apiToken string) *Client {
	return newClientUserPassAPISecure(arrayHost, username, password, apiToken,true)
}

//NewClientAPIToken New Client from API Token
func NewClientAPIToken(arrayHost, apiToken string) *Client {
	return newClientUserPassAPISecure(arrayHost, "", "", apiToken,true)
}

//NewClientUserPass New Client from User/Pass
func NewClientUserPass(arrayHost, username, password string) *Client {
	return newClientUserPassAPISecure(arrayHost, username, password, "",true)
}

//NewClient New Client from API Token
func NewClient(arrayHost string) *Client {
	return NewClientUserPass(arrayHost, "pureuser", "pureuser")
}

//NewClientUserPassInsecure New Client from User/Pass
func NewClientUserPassInsecure(arrayHost, username, password string) *Client {
	return newClientUserPassAPISecure(arrayHost, username, password, "",false)
}

//NewClientAPITokenInsecure New Client from API Token
func NewClientAPITokenInsecure(arrayHost, apiToken string) *Client {
	return newClientUserPassAPISecure(arrayHost, "", "", apiToken,true)
}

//NewClientInsecure New Client from API Token
func NewClientInsecure(arrayHost string) *Client {
	return NewClientUserPassInsecure(arrayHost, "pureuser", "pureuser")
}

//log
func (c *Client) log(level int, msg string) {
	if level <= c.LogLevel {
		fmt.Println(msg)
	}
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
