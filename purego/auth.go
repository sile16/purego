
package purego

import (
	"time"
	"fmt"
	"github.com/davecgh/go-spew/spew"
)

//PureAuthSessionV1_12 response
type PureAuthSessionV1_12 struct {
	Username string      `json:"username"`
	Msg      interface{} `json:"msg"`
}

//StartSession start session
func (c *Client) StartSession() error {
	c.sessionLock.Lock()
	defer c.sessionLock.Unlock()

	return c.startSessionUnsafe()
}

//StartSession start session
func (c *Client) startSessionUnsafe() error {
	
	c.log(3, "StartSession()")
	

	c.sessionStarted = false

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

	return err
}



//checkSession  check that we have a session and it hasn't timed out.  StartSession if needed.
func (c *Client) checkSession() error {
	c.sessionLock.Lock()  //if we lose the session we don't want 100 threads to call start session
	defer c.sessionLock.Unlock()

	//save last time in last
	last := c.lastSessionUse
	c.lastSessionUse = time.Now()  //update the last used to current time.

	if !c.sessionStarted {
	//TODO: turn this back on afterd one testing the error handling below.
	  return c.startSessionUnsafe()
	}
	
	diff := c.lastSessionUse.Sub(last)
	if diff.Minutes() > 30 {
	   return c.startSessionUnsafe()
	} 

	return nil
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
