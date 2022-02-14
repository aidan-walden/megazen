package models

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

		if fetchRes.StatusCode == http.StatusTooManyRequests {
			fmt.Println("Waiting")
			*timeouts++
			time.Sleep(time.Second * time.Duration(10*(*timeouts)))
		}
		// return nil, errors.New("Status code error: " + string(rune(fetchRes.StatusCode)) + " " + fetchRes.Status)

		return fetchRes, nil
	}
}
