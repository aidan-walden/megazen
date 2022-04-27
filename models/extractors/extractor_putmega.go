package extractors

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"megazen/models"
	"megazen/models/utils"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
)

type putmegaEntry struct {
	Extractor
}

var putmegaHost = models.Host{
	Name: "PutMega",
}

func NewPutmega(url string) *putmegaEntry {
	downloader := putmegaEntry{Extractor{originUrl: url, host: &putmegaHost}}
	return &downloader
}

func (dl *putmegaEntry) Host() *models.Host {
	return dl.host
}

func (dl *putmegaEntry) OriginUrl() string {
	return dl.originUrl
}

func (dl *putmegaEntry) Title() string {
	return dl.title
}

func (dl *putmegaEntry) parseAlbum(doc *goquery.Document, downloads *[]models.Download) error {
	if doc == nil {
		return errors.New("empty document")
	}
	title := doc.Find("title").Text()
	if title == "" {
		return errors.New("empty title")
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
			Host: dl.host,
		}

		*downloads = append(*downloads, download)
	})

	return nil
}

func (dl *putmegaEntry) parseFile(doc *goquery.Document, downloads *[]models.Download) error {
	if doc == nil {
		return errors.New("empty document")
	}
	title := doc.Find("title").Text()
	if title == "" {
		return errors.New("empty title")
	}

	button := doc.Find(".btn")

	if button.Length() == 0 {
		return errors.New("no button found")
	}

	link, found := button.Attr("href")

	if !found {
		return errors.New("no link found")
	}

	path, err := filepath.Abs("./downloads/" + title + "/" + filepath.Base(link))

	if err != nil {
		return err
	}

	savePath, err := url.QueryUnescape(path)

	if err != nil {
		return err
	}

	*downloads = append(*downloads, models.Download{
		Url:  link,
		Path: savePath,
		Host: dl.host,
	})

	return nil
}

func (dl *putmegaEntry) ParseDownloads(c chan *[]models.Download) error {
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

	if strings.Contains(dl.originUrl, "/album/") {
		err := dl.parseAlbum(doc, &downloads)
		if err != nil {
			return err
		}
	} else {
		err := dl.parseFile(doc, &downloads)
		if err != nil {
			return err
		}
	}

	return nil
}
