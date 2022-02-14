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
	Extractor
	token string
}

func NewGofile(url string, token string) *gofileEntry {
	downloader := gofileEntry{Extractor{originUrl: url, host: models.Host{Name: "Gofile"}}, token}
	return &downloader
}

func (dl *gofileEntry) Host() *models.Host {
	return &dl.host
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

	return nil
}
