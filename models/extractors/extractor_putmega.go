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

type putmegaEntry struct {
	host    models.Host
	baseUrl string
	title   string
	models.FileHostEntry
}

func NewPutmega(url string) *putmegaEntry {
	downloader := putmegaEntry{baseUrl: url, host: models.Host{
		Name: "PutMega",
	}}
	return &downloader
}

func (dl *putmegaEntry) Host() *models.Host {
	return &dl.host
}

func (dl *putmegaEntry) BaseUrl() string {
	return dl.baseUrl
}

func (dl *putmegaEntry) Title() string {
	return dl.title
}

func (dl *putmegaEntry) ParseDownloads(c chan *[]models.Download) error {
	downloads := make([]models.Download, 0)
	defer func() {
		c <- &downloads
	}()

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

	dl.title = strings.TrimSpace(doc.Find(".text-overflow-ellipsis > a:nth-child(1)").Text())
	fmt.Println("Title: ", dl.title)

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
			Host: &dl.host,
		}

		downloads = append(downloads, download)
	})

	return nil
}
