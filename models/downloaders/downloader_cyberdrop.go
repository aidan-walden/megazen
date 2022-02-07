package downloaders

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"megazen/models"
	"path/filepath"
	"strings"
)

type cyberdropDownloader struct {
	models.Host
	baseUrl string
	title   string
}

func NewCyberdrop(url string) *cyberdropDownloader {
	downloader := cyberdropDownloader{baseUrl: url, Host: models.Host{
		Name: "Cyberdrop",
	}}
	return &downloader
}

func (dl *cyberdropDownloader) ParseDownloads(c chan *[]models.Download) error {
	res, err := WaitForSuccessfulRequest(dl.baseUrl, &dl.Timeouts)

	if err != nil {
		return err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(res.Body)

	doc, err := goquery.NewDocumentFromReader(res.Body)

	if err != nil {
		return err
	}

	dl.title = strings.TrimSpace(doc.Find("#title").Text())
	fmt.Println("Title: ", dl.title)

	var downloads []models.Download

	doc.Find("a.image").Each(func(i int, s *goquery.Selection) {
		link, found := s.Attr("href")

		if !found {
			return
		}

		fileTitle := filepath.Base(link)

		savePath, err := filepath.Abs("./downloads/" + dl.title + "/" + fileTitle)

		if err != nil {
			panic(err)
		}

		download := models.Download{
			Url:  link,
			Path: savePath,
			Host: &dl.Host,
		}

		downloads = append(downloads, download)
	})

	c <- &downloads

	return nil
}
