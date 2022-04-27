package extractors

import (
	"errors"
	"fmt"
	"io"
	"megazen/models"
	"megazen/models/utils"
	"mime"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var kemonoReg = regexp.MustCompile(`^(http)s?://(?:www\.)?(kemono|coomer)\.party/`)

type kemonoEntry struct {
	Extractor
}

var kemonoHost = models.Host{
	Name: "Kemono",
}

func NewKemono(url string) *kemonoEntry {
	downloader := kemonoEntry{Extractor{originUrl: url, host: &kemonoHost}}
	return &downloader
}

func (dl *kemonoEntry) Host() *models.Host {
	return dl.host
}

func (dl *kemonoEntry) OriginUrl() string {
	return dl.originUrl
}

func (dl *kemonoEntry) Title() string {
	return dl.title
}

func (dl *kemonoEntry) ParseDownloads(c chan *[]models.Download) error {
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

		savePath := filepath.Join(dl.title+"/", fileTitle)

		downloads = append(downloads, models.Download{
			Url:  link,
			Path: savePath,
			Host: dl.host,
		})
	})

	return nil
}
