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
	Extractor
}

func NewCyberdrop(url string) *cyberdropEntry {
	downloader := cyberdropEntry{Extractor{originUrl: url, host: models.Host{
		Name: "Cyberdrop",
	}}}
	return &downloader
}

func (dl *cyberdropEntry) Host() *models.Host {
	return &dl.host
}

func (dl *cyberdropEntry) OriginUrl() string {
	return dl.originUrl
}

func (dl *cyberdropEntry) Title() string {
	return dl.title
}

func (dl *cyberdropEntry) ParseDownloads(c chan *[]models.Download) error {
	downloads := make([]models.Download, 0)
	defer func() {
		c <- &downloads
	}()

	res, err := models.WaitForSuccessfulRequest(dl.originUrl, &dl.host.Timeouts)

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

	return nil
}
