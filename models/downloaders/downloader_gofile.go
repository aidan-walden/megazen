package downloaders

import (
	"encoding/json"
	"io"
	"megazen/models"
	"megazen/models/gofile"
	"net/url"
	"path/filepath"
)

type gofileDownloader struct {
	baseURL string
	name    string
	token   string
	models.Host
}

func NewGofile(url string, token string) *gofileDownloader {
	downloader := gofileDownloader{baseURL: url, token: token, Host: models.Host{Name: "Gofile"}}
	return &downloader
}

func (dl *gofileDownloader) ParseDownloads(c chan *[]models.Download) error {
	contentId := filepath.Base(dl.baseURL)

	res, err := WaitForSuccessfulRequest("https://apiv2.gofile.io/getContent?contentId="+contentId+"&token="+dl.token+"&websiteToken=websiteToken&cache=true", &dl.Timeouts)

	if err != nil {
		return err
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

	dl.name = content.Data.Name

	for _, file := range content.Data.Contents {

		fileTitle, err := url.QueryUnescape(filepath.Base(file.Link))

		if err != nil {
			panic(err)
		}

		savePath, err := filepath.Abs("./downloads/" + dl.name + "/" + fileTitle)

		if err != nil {
			panic(err)
		}

		headers := make(map[string]string)
		headers["Cookie"] = "accountToken=" + dl.token

		dl.Headers = &headers

		downloads = append(downloads, models.Download{
			Url:  file.Link,
			Path: savePath,
			Host: &dl.Host,
		})
	}

	c <- &downloads

	return nil
}
