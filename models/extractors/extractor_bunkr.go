package extractors

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"megazen/models"
	"mime"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
)

type bunkrEntry struct {
	Extractor
}

func NewBunkr(url string) *bunkrEntry {
	downloader := bunkrEntry{Extractor{originUrl: url, host: models.Host{
		Name: "Bunkr",
	}}}
	return &downloader
}

func (dl *bunkrEntry) Host() *models.Host {
	return &dl.host
}

func (dl *bunkrEntry) OriginUrl() string {
	return dl.originUrl
}

func (dl *bunkrEntry) Title() string {
	return dl.title
}

func (dl *bunkrEntry) ParseDownloads(c chan *[]models.Download) error {
	downloads := make([]models.Download, 0)
	defer func() {
		c <- &downloads
	}()

	if strings.Contains(dl.originUrl, "stream.bunkr.is") {

		path, err := filepath.Abs("./downloads/" + filepath.Base(dl.originUrl))

		if err != nil {
			return err
		}

		savePath, err := url.QueryUnescape(path)

		if err != nil {
			return err
		}

		downloads = append(downloads, models.Download{
			Url:  strings.Replace(dl.originUrl, "stream.bunkr.is/v/", "stream.bunkr.is/d/", 1),
			Path: savePath,
			Host: &dl.host,
		})
		return nil
	}

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

		downloads = append(downloads, models.Download{
			Url:  link,
			Path: savePath,
			Host: &dl.host,
		})
	})

	return nil
}
