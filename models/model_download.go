package models

import (
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"
)

type Download struct {
	Url      string
	Host     *Host
	Path     string
	Complete bool
	total    int64   // total Expected length of the file in bytes
	current  int64   // current Amount of bytes downloaded
	Progress float64 // Progress Percentage of the download completed
	io.Reader
}

func (d *Download) Read(p []byte) (int, error) {
	n, err := d.Reader.Read(p)
	if n > 0 {
		d.current += int64(n)
		d.Progress = float64(d.current) / float64(d.total) * 100
		fmt.Printf("Progress for %s: %f\n", d.Path, d.Progress)
	}

	return n, err
}

// DownloadFile handles downloading files from their direct URLs
// and saving them to the specified path.
func (dl *Download) DownloadFile() error {
	fmt.Println(dl.Url)
	var res *http.Response

	client := &http.Client{}

	// c.L.Lock()
	for {
		// Get the data
		req, err := http.NewRequest("GET", dl.Url, nil)
		if err != nil {
			return err
		}

		if dl.Host.Headers != nil {
			for k, v := range *dl.Host.Headers {
				req.Header.Set(k, v)
			}
		}

		fetchRes, err := client.Do(req)
		if err != nil {
			return err
		}

		if fetchRes.StatusCode != 200 {
			if fetchRes.StatusCode == 404 {
				return nil
			} else if fetchRes.StatusCode == 429 {
				dl.Host.Lock.Lock()
				fmt.Println("Download Waiting, timeouts = ", dl.Host.Timeouts)
				atomic.AddInt32(&dl.Host.Timeouts, 1)
				time.Sleep(time.Duration(math.Pow(5, float64(dl.Host.Timeouts+1))) * time.Second)
				fmt.Println("Download Resuming")
				dl.Host.Lock.Unlock()
			} else {
				return errors.New("Status code error: " + string(rune(fetchRes.StatusCode)) + " " + fetchRes.Status)
			}
		} else {
			atomic.StoreInt32(&dl.Host.Timeouts, 0)
			res = fetchRes
			break
		}
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(res.Body)

	dl.Reader = res.Body
	dl.total = res.ContentLength

	// Check if this file has already been downloaded
	stat, err := os.Stat(dl.Path)

	if err == nil {
		if stat.Size() == dl.total {
			fmt.Println("File already downloaded")
			return nil
		}
	}

	// Create the file
	if err := os.MkdirAll(filepath.Dir(dl.Path), 0755); err != nil {
		return err
	}
	out, err := os.Create(dl.Path + ".tmp")
	if err != nil {
		err := out.Close()
		if err != nil {
			return err
		}
		return err
	}

	// Write the body to file
	_, err = io.Copy(out, dl)
	if err != nil {
		err := out.Close()
		if err != nil {
			return err
		}
	}

	err = out.Close()
	if err != nil {
		return err
	}

	// Rename the file
	err = os.Rename(dl.Path+".tmp", dl.Path)
	if err != nil {
		err := out.Close()
		if err != nil {
			return err
		}
		return err
	}

	dl.Complete = true
	return nil
}
