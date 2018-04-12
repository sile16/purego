package purego

import (
	"fmt"

)


//apiCall internal api call, that unmarshalls JSON into specific type.
func (c *Client) apiCall(method string, endPoint string) error {
	return c.apiCallJSON(method,endPoint,nil,nil)
}


//apiCallJSON internal api call, that unmarshalls JSON into specific type.
func (c *Client) apiCallJSON(method string, endPoint string, reqData []byte, resInt interface{}) error {
	//todo
	c.log(3, "API Call: "+endPoint)

	//this will block if channel reaches 10 concurrent threads a.k.a the channel size of c.maxAPICAlls
	c.maxAPICalls <- struct{}{}
	//defer the release of channel, by reading a value out.
	defer func() {
		<-c.maxAPICalls //release semaphore lock / clear the channel
	}()


	errSession := c.checkSession()
	if errSession != nil {
		return errSession
	}

	var resp *PureHTTPResponse
	var err error

	//provide some sort of retry loop
	for retry := 0; retry < 2; retry++ {

		resp, err = c.doHTTPRequest(method, endPoint, reqData, resInt)
		if err != nil {
			return err
		}

		switch resp.StatusCode {
		case 200:
			//Everything Okay Return
			return nil

		case 401:
			//UNAUTHORIZED
			//This means we need to start our session.
			c.log(3, "Unauthrorized, restarting session and retrying")
			err = c.StartSession()

			if err != nil {
				return err
			}
			//then retry
			//let loop circle back to top
			break

		//case 400:
			//BAD Request
			//invalid action or invalid o rmissing data 
			//No reason to send another bad request
			//return fmt.Errorf("Error: %d  Response: %s", resp.StatusCode, resp.Body)
		
		//case 404:
			//Not Found
			//URI does not exist
		
		//case 403:
			//Forbidden
			//valid request but user is not allowed.
			//return fmt.Errorf("Error: %d  Response: %s", resp.StatusCode, resp.Body)

		default:
			//This will capture the list above and all other status codes.
			//all other error conditions
			return fmt.Errorf("Error: %d  Response: %s", resp.StatusCode, resp.Body)
		}

	}

	//Hit max retry limit, 
	return fmt.Errorf("Max retry hit")

}
