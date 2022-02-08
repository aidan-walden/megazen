package extractors

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"megazen/models"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
)

type cyberdropEntry struct {
	host    models.Host
	baseUrl string
	title   string
	models.FileHostEntry
}

func NewCyberdrop(url string) *cyberdropEntry {
	downloader := cyberdropEntry{baseUrl: url, host: models.Host{
		Name: "Cyberdrop",
	}}
	return &downloader
}

func (dl *cyberdropEntry) Host() *models.Host {
	return &dl.host
}

func (dl *cyberdropEntry) BaseUrl() string {
	return dl.baseUrl
}

func (dl *cyberdropEntry) Title() string {
	return dl.title
}

func (dl *cyberdropEntry) ParseDownloads(c chan *[]models.Download) error {
	res, err := models.WaitForSuccessfulRequest(dl.baseUrl, &dl.host.Timeouts)

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
			Host: &dl.host,
		}

		downloads = append(downloads, download)
	})

	c <- &downloads

	return nil
}
