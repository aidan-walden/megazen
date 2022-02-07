package downloaders

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

func WaitForSuccessfulRequest(url string, timeouts *int32) (*http.Response, error) {
	for {
		fetchRes, err := http.Get(url)
		if err != nil {
			return nil, err
		}

		if fetchRes.StatusCode != 200 {
			if fetchRes.StatusCode == 404 {
				return nil, nil
			} else if fetchRes.StatusCode == 429 {
				fmt.Println("Waiting")
				*timeouts++
				time.Sleep(time.Second * time.Duration(10*(*timeouts)))
			} else {
				return nil, errors.New("Status code error: " + string(rune(fetchRes.StatusCode)) + " " + fetchRes.Status)
			}
		}

		return fetchRes, nil
	}
}
