package downloaders

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"megazen/models"
	"net/url"
	"path/filepath"
	"strings"
)

type putmegaDownloader struct {
	models.Host
	baseUrl string
	title   string
}

func NewPutmega(url string) *putmegaDownloader {
	downloader := putmegaDownloader{baseUrl: url, Host: models.Host{
		Name: "PutMega",
	}}
	return &downloader
}

func (dl *putmegaDownloader) ParseDownloads(c chan *[]models.Download) error {
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

	dl.title = strings.TrimSpace(doc.Find(".text-overflow-ellipsis > a:nth-child(1)").Text())
	fmt.Println("Title: ", dl.title)

	var downloads []models.Download

	doc.Find("a.image-container").Each(func(i int, s *goquery.Selection) {
		link, found := s.Find("img").Attr("src")

		if !found {
			return
		}

		link = strings.Replace(link, ".md.", ".", 1)

		fileTitle, err := url.QueryUnescape(filepath.Base(link))

		if err != nil {
			panic(err)
		}

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
