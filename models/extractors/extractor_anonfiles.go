package extractors

import (
	"errors"
	"github.com/PuerkitoBio/goquery"
	"io"
	"megazen/models"
	"net/http"
	"net/url"
	"path/filepath"
)

type anonfilesEntry struct {
	Extractor
}

func NewAnonfiles(url string) *anonfilesEntry {
	downloader := anonfilesEntry{Extractor{originUrl: url, host: models.Host{
		Name: "AnonFiles",
	}}}
	return &downloader
}

func (dl *anonfilesEntry) Host() *models.Host {
	return &dl.host
}

func (dl *anonfilesEntry) OriginUrl() string {
	return dl.originUrl
}

func (dl *anonfilesEntry) Title() string {
	return dl.title
}

func (dl *anonfilesEntry) ParseDownloads(c chan *[]models.Download) error {
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

	downloadBtn := doc.Find("#download-url")
	link, found := downloadBtn.Attr("href")
	if !found {
		return nil
	}

	title, err := url.QueryUnescape(filepath.Base(link))

	if err != nil {
		return err
	}

	dl.title = title

	savePath, err := filepath.Abs("./downloads/" + dl.title)

	headers := make(map[string]string)
	headers["Referer"] = "https://anonfiles.com/"

	dl.host.Headers = &headers

	downloads = append(downloads, models.Download{
		Url:  link,
		Path: savePath,
		Host: &dl.host,
	})

	return nil
}
