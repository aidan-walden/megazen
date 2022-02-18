package models

import (
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sync/atomic"
	"time"
)

// Download represents one direct download link
type Download struct {
	Url      string
	Host     *Host
	Path     string
	complete bool
	total    int64   // total Expected length of the file in bytes
	current  int64   // current Amount of bytes downloaded
	progress float64 // Progress Percentage of the download completed
	errors   []error
	io.Reader
}

type DownloadResponse struct {
	Url      string  `json:"url"`
	Path     string  `json:"path"`
	Complete bool    `json:"complete"`
	Total    int64   `json:"total"`    // total Expected length of the file in bytes
	Current  int64   `json:"current"`  // current Amount of bytes downloaded
	Progress float64 `json:"progress"` // Progress Percentage of the download completed
}

type DownloadSubmission struct {
	Url      string `json:"url"`
	Password string `json:"password"`
}

// FileHostEntry Represents one link from a generic file host
type FileHostEntry interface {
	Host() *Host
	OriginUrl() string
	Title() string
	ParseDownloads(c chan *[]Download) error
}

func (dl *Download) Read(p []byte) (int, error) {
	n, err := dl.Reader.Read(p)
	if n > 0 {
		dl.current += int64(n)
		dl.progress = float64(dl.current) / float64(dl.total) * 100
	}

	return n, err
}

func (dl *Download) Progress() float64 {
	return dl.progress
}

func (dl *Download) IsComplete() bool {
	return dl.complete
}

func (dl *Download) Total() int64 {
	return dl.total
}

func (dl *Download) Current() int64 {
	return dl.current
}

func (dl *Download) Errors() []error {
	return dl.errors
}

var re = regexp.MustCompile("[|&;$%@\"<>()+,?]")

// DownloadFile handles downloading files from their direct URLs
// and saving them to the specified path.
func (dl *Download) DownloadFile() error {
	fmt.Println(dl.Url)

	defer func() {
		dl.complete = true
	}()

	var res *http.Response

	// Make file path valid
	dl.Path = re.ReplaceAllString(dl.Path, "-")
	fmt.Println("Downloading to:", dl.Path)

	client := &http.Client{}

	// c.L.Lock()
	for {
		// Get the data
		req, err := http.NewRequest("GET", dl.Url, nil)
		req.Close = true
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
			dl.errors = append(dl.errors, err)
			return err
		}

		if fetchRes.StatusCode != 200 {
			if fetchRes.StatusCode == 404 {
				return nil
			} else if fetchRes.StatusCode == 429 {
				dl.Host.Lock.Lock()
				fmt.Println("Download Waiting, timeouts = ", dl.Host.Timeouts)
				atomic.AddInt32(&dl.Host.Timeouts, 1)
				time.Sleep(time.Duration(math.Max(math.Pow(5, float64(dl.Host.Timeouts+1)), 10)) * time.Second)
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

	dl.Reader = res.Body
	dl.total = res.ContentLength

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			dl.errors = append(dl.errors, err)
			fmt.Println("Error closing body:", err)
		}
	}(res.Body)

	// Check if this file has already been downloaded
	stat, err := os.Stat(dl.Path)

	if err == nil {
		if stat.Size() == dl.total {
			fmt.Println("File already downloaded")
			dl.complete = true
			return nil
		} else {
			dl.errors = append(dl.errors, err)
		}
	}

	// Create the file
	if err := os.MkdirAll(filepath.Dir(dl.Path), 0755); err != nil {
		return err
	}
	out, err := os.Create(dl.Path + ".tmp")
	if err != nil {
		dl.errors = append(dl.errors, err)
		err := out.Close()
		if err != nil {
			dl.errors = append(dl.errors, err)
			return err
		}
		return err
	}

	// Write the body to file
	_, err = io.Copy(out, dl)
	if err != nil {
		dl.errors = append(dl.errors, err)
		err := out.Close()
		if err != nil {
			dl.errors = append(dl.errors, err)
			return err
		}
		return err
	}

	err = out.Close()
	if err != nil {
		dl.errors = append(dl.errors, err)
		return err
	}

	// Rename the file
	err = os.Rename(dl.Path+".tmp", dl.Path)
	if err != nil {
		dl.errors = append(dl.errors, err)
		return err
	}

	fmt.Println(dl.Path + ": Download complete")

	return nil
}
