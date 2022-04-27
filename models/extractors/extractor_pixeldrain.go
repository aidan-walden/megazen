package extractors

import (
	"encoding/json"
	"errors"
	"megazen/models"
	"megazen/models/pixeldrain"
	"megazen/models/utils"
	"net/http"
	"path/filepath"
)

type pixeldrainEntry struct {
	Extractor
	isFolder bool
}

var pixelHost = models.Host{
	Name: "Pixeldrain",
}

func NewPixeldrain(url string, isFolder bool) *pixeldrainEntry {
	downloader := pixeldrainEntry{Extractor{originUrl: url, host: &pixelHost}, isFolder}
	return &downloader
}

func (dl *pixeldrainEntry) Host() *models.Host {
	return dl.host
}

func (dl *pixeldrainEntry) OriginUrl() string {
	return dl.originUrl
}

func (dl *pixeldrainEntry) Title() string {
	return dl.title
}

func (dl *pixeldrainEntry) ParseDownloads(c chan *[]models.Download) error {
	downloads := make([]models.Download, 0)
	defer func() {
		c <- &downloads
	}()

	fileId := filepath.Base(dl.originUrl)

	if dl.isFolder {
		res, err := utils.WaitForSuccessfulRequest("https://pixeldrain.com/api/list/"+fileId, &dl.host.Timeouts)

		if err != nil {
			return err
		}

		if res.StatusCode != http.StatusOK {
			return errors.New("Status code error: " + string(rune(res.StatusCode)) + " " + res.Status)
		}

		var folder pixeldrain.FolderInfo

		err = json.NewDecoder(res.Body).Decode(&folder)

		if err != nil {
			return err
		}

		for _, info := range folder.Files {
			savePath, err := filepath.Abs("./downloads/" + folder.Title + "/" + info.Name)

			if err != nil {
				panic(err)
			}

			downloads = append(downloads, models.Download{
				Url:  "https://pixeldrain.com/api/file/" + info.Id,
				Path: savePath,
				Host: dl.host,
			})
		}

	} else {
		res, err := utils.WaitForSuccessfulRequest("https://pixeldrain.com/api/file/"+fileId+"/info", &dl.host.Timeouts)

		if err != nil {
			return err
		}

		if res.StatusCode != http.StatusOK {
			return errors.New("Status code error: " + string(rune(res.StatusCode)) + " " + res.Status)
		}

		var info pixeldrain.FileInfo

		err = json.NewDecoder(res.Body).Decode(&info)

		if err != nil {
			return err
		}

		downloads = append(downloads, models.Download{
			Url:  "https://pixeldrain.com/api/file/" + info.Id,
			Path: info.Name,
			Host: dl.host,
		})
	}

	return nil
}
