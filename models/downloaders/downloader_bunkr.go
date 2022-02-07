package downloaders

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"megazen/models"
	"mime"
	"net/url"
	"path/filepath"
	"strings"
)

type bunkrDownloader struct {
	models.Host
	baseUrl string
	title   string
}

func NewBunkr(url string) *bunkrDownloader {
	downloader := bunkrDownloader{baseUrl: url, Host: models.Host{
		Name: "Bunkr",
	}}
	return &downloader
}

func (dl *bunkrDownloader) ParseDownloads(c chan *[]models.Download) error {
	if strings.Contains(dl.baseUrl, "stream.bunkr.is") {

		path, err := filepath.Abs("./downloads/" + filepath.Base(dl.baseUrl))

		if err != nil {
			return err
		}

		savePath, err := url.QueryUnescape(path)

		if err != nil {
			return err
		}

		replaced := make([]models.Download, 1)
		replaced[0] = models.Download{
			Url:  strings.Replace(dl.baseUrl, "stream.bunkr.is/v/", "stream.bunkr.is/d/", 1),
			Path: savePath,
			Host: &dl.Host,
		}
		c <- &replaced
		return nil
	}

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

	doc.Find(".image-container.column").Each(func(i int, s *goquery.Selection) {
		link, found := s.Find("a").First().Attr("href")

		if !found {
			return
		}

		extension := filepath.Ext(link)

		mediaType := mime.TypeByExtension(extension)

		if strings.HasPrefix(mediaType, "image") {
			link = strings.Replace(link, "cdn.bunkr.is", "i.bunkr.is", 1)
		} else if strings.HasPrefix(mediaType, "video") {
			link = strings.Replace(link, "cdn.bunkr.is", "stream.bunkr.is/d", 1)
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
