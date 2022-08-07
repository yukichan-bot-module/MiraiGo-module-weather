package pkg

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

// HTTPGetRequest send a http get request
func HTTPGetRequest(url string, queryList [][]string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	for _, queryItem := range queryList {
		if len(queryItem) != 2 {
			return nil, fmt.Errorf("%v is not a valid query item", queryItem)
		}
		q.Add(queryItem[0], queryItem[1])
	}
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
