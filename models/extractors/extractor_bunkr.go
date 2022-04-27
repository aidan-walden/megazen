package extractors

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"megazen/models"
	"megazen/models/bunkr"
	"megazen/models/utils"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/panjf2000/ants"
)

type bunkrEntry struct {
	Extractor
}

var streamBunkr = regexp.MustCompile(`((stream|cdn)\.bunkr\.[a-z]+)((/[a-z]/)*)`)
var albumBunkr = regexp.MustCompile(`/a/[a-zA-Z0-9]+`)

var bunkrPool, _ = ants.NewPool(2)

var bunkrHost = models.Host{
	Name: "Bunkr",
	Pool: bunkrPool,
	Wg:   &sync.WaitGroup{},
}

func NewBunkr(url string) *bunkrEntry {
	downloader := bunkrEntry{Extractor{originUrl: url, host: &bunkrHost}}
	return &downloader
}

func (dl *bunkrEntry) Host() *models.Host {
	return dl.host
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

	dl.originUrl = strings.Replace(dl.originUrl, "bunkr.to", "bunkr.is", 1)

	if strings.Contains(dl.originUrl, "stream.bunkr.is") || strings.Contains(dl.originUrl, "cdn.bunkr.is") {

		path, err := filepath.Abs("./downloads/" + filepath.Base(dl.originUrl))

		if err != nil {
			return err
		}

		savePath, err := url.QueryUnescape(path)

		if err != nil {
			return err
		}

		downloads = append(downloads, models.Download{
			Url:  streamBunkr.ReplaceAllString(dl.originUrl, "media-files.bunkr.is/"),
			Path: savePath,
			Host: dl.host,
		})
		return nil
	}

	albumPath := albumBunkr.FindString(dl.originUrl)

	fmt.Println(albumPath)

	res, err := utils.WaitForSuccessfulRequest("https://bunkr.is"+strings.Replace(albumPath, "/a/", "/api/album/", 1), &dl.host.Timeouts)

	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return errors.New("Status code error: " + string(rune(res.StatusCode)) + " " + res.Status + ". Caused by WaitForSuccessfulRequest")
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(res.Body)

	var album bunkr.AlbumResponse

	err = json.NewDecoder(res.Body).Decode(&album)

	if err != nil {
		return err
	}

	dl.title = album.Name

	fmt.Println("Got album: " + album.Name)

	for _, file := range album.Files {
		savePath, err := filepath.Abs("./downloads/" + dl.title + "/" + file.Name)

		if err != nil {
			panic(err)
		}

		downloads = append(downloads, models.Download{
			Url:  streamBunkr.ReplaceAllString(file.Url, "media-files.bunkr.is"),
			Path: savePath,
			Host: dl.host,
		})
	}

	return nil
}
