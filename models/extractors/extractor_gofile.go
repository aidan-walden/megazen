package extractors

import (
	"encoding/json"
	"errors"
	"io"
	"megazen/models"
	"megazen/models/gofile"
	"net/http"
	"net/url"
	"path/filepath"
)

type gofileEntry struct {
	host    models.Host
	baseUrl string
	title   string
	token   string
	models.FileHostEntry
}

func NewGofile(url string, token string) *gofileEntry {
	downloader := gofileEntry{baseUrl: url, token: token, host: models.Host{Name: "Gofile"}}
	return &downloader
}

func (dl *gofileEntry) Host() *models.Host {
	return &dl.host
}

func (dl *gofileEntry) BaseUrl() string {
	return dl.baseUrl
}

func (dl *gofileEntry) Title() string {
	return dl.title
}

func (dl *gofileEntry) ParseDownloads(c chan *[]models.Download) error {
	contentId := filepath.Base(dl.baseUrl)

	res, err := models.WaitForSuccessfulRequest("https://apiv2.gofile.io/getContent?contentId="+contentId+"&token="+dl.token+"&websiteToken=websiteToken&cache=true", &dl.host.Timeouts)

	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return errors.New("Status code error: " + string(rune(res.StatusCode)) + " " + res.Status)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(res.Body)

	var downloads []models.Download

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

		headers := make(map[string]string)
		headers["Cookie"] = "accountToken=" + dl.token

		dl.host.Headers = &headers

		downloads = append(downloads, models.Download{
			Url:  file.Link,
			Path: savePath,
			Host: &dl.host,
		})
	}

	c <- &downloads

	return nil
}
