
//Package purego Wrapper for the Pure Storage API
package purego

import (
	"time"
	
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

	authStatus int
	retryCount int
	retryMax int

	apiVersion string
	httpClient *http.Client
	cookieJar *cookiejar.Jar
	sessionStarted bool
	LogLevel int //0 - error, 1 - Warning, 2 - Info, 3 - Debug
	
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

//log
func (c *Client) log(level int,msg string) () {
	if level <= c.LogLevel{
		fmt.Println(msg)
	}
}


func (c *Client) init() {
	c.apiVersion = "1.12"
	c.url = "https://"+c.arrayHost+"/api/" + c.apiVersion + "/"
	c.cookieJar,_ = cookiejar.New(nil)
	c.httpClient = &http.Client{Jar: c.cookieJar}
	c.sessionStarted = false
	c.LogLevel = 3

	//initially set timeout small so that network, auth happen quick or fail quick
	//After successful session we will raise timeout.
	c.httpClient.Timeout = 3 * time.Second
	
	if c.APIToken == "" {
		c.StartSession()
	}
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
}

//doHTTPRequest internal api call
func (c *Client) doHTTPRequest(method string,endPoint string, reqData []byte, resInt interface{}  ) ( *http.Response, error ) {
	//append endpoint to base url
	c.log(3,"doHTTP Request: "+endPoint)
	url := c.url + endPoint
	//fmt.Println(url)

	//setup request
	req, err := http.NewRequest(method,url,bytes.NewBuffer(reqData))
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
		return nil,err
	}

	//Unmarshal the response based on the passed interface
	if resInt != nil {
		json.Unmarshal(body,&resInt)
	}
	return resp,nil
}

//apiCall internal api call
func (c *Client) apiCall(method string,endPoint string, reqData []byte, resInt interface{}  ) ( error ) {
	c.log(3,"API Call: "+endPoint)

	if ! c.sessionStarted {
		//TODO: turn this back on afterd one testing the error handling below.
		err := c.StartSession()
		if err != nil {
			return err
		}
	}
	
	resp,err := c.doHTTPRequest(method,endPoint,reqData,resInt)
	if err != nil {
		return err
	}

	for retry := 0 ; retry <2; retry++ {
		
		switch resp.StatusCode {
		case 401:
			//UNAUTHORIZED
			//This means we need to start our session.
			c.log(3,"Unauthrorized, restarting session and retrying")
			c.sessionStarted = false
			err=c.StartSession()
			if err != nil{
				return err
			}
			//then retry
			resp,err = c.doHTTPRequest(method,endPoint,reqData,resInt)
			break
			
	
		case 200:
			//Everything Okay Return
			return nil

		default:
			//all other error conditions
			body, _ := ioutil.ReadAll(resp.Body)
			return fmt.Errorf("Error: %d  Response: %s",resp.StatusCode, string(body))
		}
		
	}

	if err != nil || resp.StatusCode != 200 {
		return fmt.Errorf("Error: Max retry hit last error: %s",err.Error())
	}
	
	return nil
	
}

//PureAuthSessionV1_12 response
type PureAuthSessionV1_12 struct {
	Username string `json:"username"`
}

//StartSession start session
func (c *Client) StartSession () error {
	c.log(3,"StartSession()")
	data := []byte(`{"api_token": "` + c.APIToken + `"}`)
	result := PureAuthSessionV1_12{}

	_, err := c.doHTTPRequest("POST","auth/session",data,&result)
	
	if err != nil {
		fmt.Println(err)
		return err
	}

	if result.Username == "" {
		//error no matching user name found
		return fmt.Errorf("Error no username returned when starting session")
	}

	c.log(3,"Started session as user: "+result.Username)
	c.sessionStarted = true
	//Some large API calls make take some time.
	c.httpClient.Timeout = 20 * time.Second
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
	c.log(3,"Get Array")
	result := PureArrayV1_12{}
	err := c.apiCall("GET","array",nil,&result)
	if err != nil {
		c.log(0,err.Error())
	}
	//c.log(4,string(result))
	return result
}

//PureVolumesV1_12 d
type PureVolumesV1_12 []struct {
	Total             int64       `json:"total,omitempty"`
	Name              string      `json:"name"`
	System            int64       `json:"system,omitempty"`
	Snapshots         int64       `json:"snapshots,omitempty"`
	Volumes           int64       `json:"volumes,omitempty"`
	DataReduction     float64     `json:"data_reduction,omitempty"`
	Size              int64       `json:"size,omitempty"`
	SharedSpace       int64       `json:"shared_space,omitempty"`
	ThinProvisioning  float64     `json:"thin_provisioning,omitempty"`
	TotalReduction    float64     `json:"total_reduction,omitempty"`
	WritesPerSec      int         `json:"writes_per_sec,omitempty"`
	UsecPerWriteOp    int         `json:"usec_per_write_op,omitempty"`
	OutputPerSec      int         `json:"output_per_sec,omitempty"`
	SanUsecPerReadOp  int         `json:"san_usec_per_read_op,omitempty"`
	ReadsPerSec       int         `json:"reads_per_sec,omitempty"`
	InputPerSec       int         `json:"input_per_sec,omitempty"`
	Time              time.Time   `json:"time,omitempty"`
	SanUsecPerWriteOp int         `json:"san_usec_per_write_op,omitempty"`
	UsecPerReadOp     int         `json:"usec_per_read_op,omitempty"`
	Source            string      `json:"source,omitempty"`
	Serial            string      `json:"serial,omitempty"`
	Created           time.Time   `json:"created,omitempty"`
}

//GetVolumes tesitn
func (c *Client) GetVolumes() PureVolumesV1_12 {
	c.log(3,"GetVolumes)")
	result := PureVolumesV1_12{}
	err := c.apiCall("GET","volume",nil,&result)
	if err != nil {
		c.log(0,err.Error())
	}
	//c.log(4,string(result))
	return result
}