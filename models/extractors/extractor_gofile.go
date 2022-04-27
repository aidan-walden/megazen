package extractors

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"megazen/models"
	"megazen/models/gofile"
	"megazen/models/utils"
	"net/http"
	"net/url"
	"path/filepath"
)

type gofileEntry struct {
	Extractor
	token    string
	password string
}

var gofileHost = models.Host{Name: "Gofile"}

func NewGofile(url string, token string, password string) *gofileEntry {
	downloader := gofileEntry{Extractor{originUrl: url, host: &gofileHost}, token, password}
	downloader.host.Cookies = make([]*http.Cookie, 1)
	downloader.host.Cookies[0] = &http.Cookie{Name: "accountToken", Value: token, MaxAge: 0}

	return &downloader
}

func (dl *gofileEntry) Host() *models.Host {
	return dl.host
}

func (dl *gofileEntry) OriginUrl() string {
	return dl.originUrl
}

func (dl *gofileEntry) Title() string {
	return dl.title
}

func (dl *gofileEntry) ParseDownloads(c chan *[]models.Download) error {
	downloads := make([]models.Download, 0)
	defer func() {
		c <- &downloads
	}()

	contentId := filepath.Base(dl.originUrl)

	requestUrl := "https://apiv2.gofile.io/getContent?contentId=" + contentId + "&token=" + dl.token + "&websiteToken=12345&cache=true"
	if dl.password != "" {
		sha256Password := sha256.Sum256([]byte(dl.password))
		requestUrl += "&password=" + hex.EncodeToString(sha256Password[:])
	}

	res, err := utils.WaitForSuccessfulRequest(requestUrl, &dl.host.Timeouts)

	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return errors.New("Status code error: " + string(rune(res.StatusCode)) + " " + res.Status)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(res.Body)

	var content gofile.Content

	err = json.NewDecoder(res.Body).Decode(&content)
	if err != nil {
		return err
	}

	dl.title = content.Data.Name

	for _, file := range content.Data.Contents {

		fileTitle, err := url.QueryUnescape(filepath.Base(file.Link))

		if err != nil {
			panic(err)
		}

		savePath, err := filepath.Abs("./downloads/" + dl.title + "/" + fileTitle)

		if err != nil {
			panic(err)
		}

		downloads = append(downloads, models.Download{
			Url:  file.Link,
			Path: savePath,
			Host: dl.host,
		})
	}

	return nil
}
