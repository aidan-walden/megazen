package extractors

import (
	"errors"
	"fmt"
	"io"
	"megazen/models"
	"megazen/models/utils"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type cyberdropEntry struct {
	Extractor
}

var cyberHost = models.Host{
	Name: "Cyberdrop",
}

func NewCyberdrop(url string) *cyberdropEntry {
	downloader := cyberdropEntry{Extractor{originUrl: url, host: &cyberHost}}
	return &downloader
}

func (dl *cyberdropEntry) Host() *models.Host {
	return dl.host
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

	res, err := utils.WaitForSuccessfulRequest(dl.originUrl, &dl.host.Timeouts)

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
	dl.title = utils.ValidTitleString(dl.title)
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

		savePath := filepath.Join(dl.title + "/" + fileTitle)

		download := models.Download{
			Url:  link,
			Path: savePath,
			Host: dl.host,
		}

		downloads = append(downloads, download)
	})

	return nil
}
