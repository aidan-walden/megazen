package extractors

import (
	"megazen/models"
	"net/url"
	"path/filepath"
)

type directEntry struct {
	Extractor
}

var directHost = models.Host{
	Name: "< Direct Download >",
}

func NewDirect(url string) *directEntry {
	downloader := directEntry{Extractor{originUrl: url, host: &directHost}}
	return &downloader
}

func (dl *directEntry) Host() *models.Host {
	return dl.host
}

func (dl *directEntry) OriginUrl() string {
	return dl.originUrl
}

func (dl *directEntry) Title() string {
	return dl.title
}

func (dl *directEntry) SetTitle(title string) {
	dl.title = title
}

func (dl *directEntry) ParseDownloads(c chan *[]models.Download) error {
	downloads := make([]models.Download, 1)

	path, err := filepath.Abs("./downloads/" + dl.title + "/" + filepath.Base(dl.originUrl))

	if err != nil {
		return err
	}

	savePath, err := url.QueryUnescape(path)

	if err != nil {
		return err
	}

	headers := make(map[string]string)
	headers["User-Agent"] = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:97.0) Gecko/20100101 Firefox/97.0"

	dl.host.Headers = &headers

	downloads[0] = models.Download{
		Url:                   dl.originUrl,
		Path:                  savePath,
		Host:                  dl.host,
		UseContentDisposition: true,
	}

	defer func() {
		c <- &downloads
	}()

	return nil
}
