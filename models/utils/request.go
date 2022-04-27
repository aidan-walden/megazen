package utils

import (
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

		if fetchRes.StatusCode == http.StatusTooManyRequests || fetchRes.StatusCode == http.StatusForbidden {
			fmt.Println("Waiting")
			*timeouts++
			time.Sleep(time.Second * time.Duration(10*(*timeouts)))
		}
		// return nil, errors.New("Status code error: " + string(rune(fetchRes.StatusCode)) + " " + fetchRes.Status)
		if fetchRes.StatusCode == http.StatusOK {
			return fetchRes, nil
		}
	}
}
