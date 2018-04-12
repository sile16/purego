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
	c.log(3, "API Call: "+endPoint)

	errSession := c.checkSession()
	if errSession != nil {
		return errSession
	}

	resp, err := c.doHTTPRequest(method, endPoint, reqData, resInt)
	if err != nil {
		return err
	}

	for retry := 0; retry < 2; retry++ {

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
			resp, err = c.doHTTPRequest(method, endPoint, reqData, resInt)
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

	//Do one last check to see if there was an error
	if err != nil || resp.StatusCode != 200 {
		return fmt.Errorf("Error: Max retry hit last error: %s", err.Error())
	}

	//Everything looks okay return nil.
	return nil

}
