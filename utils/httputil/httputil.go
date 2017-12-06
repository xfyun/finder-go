package httputil

import (
	"net/http"
	"fmt"
	"io/ioutil"
)

func DoGet(client *http.Client, url string) ([]byte, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Println("status_code:", resp.StatusCode)
	}

	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = resp.Body.Close()
	if err != nil {

	}

	return result, nil
}
