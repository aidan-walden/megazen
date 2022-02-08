package extractors

import (
	"encoding/json"
	"megazen/models"
	"megazen/models/pixeldrain"
	"path/filepath"
)

type pixeldrainEntry struct {
	host     models.Host
	baseUrl  string
	title    string
	isFolder bool
	models.FileHostEntry
}

func NewPixeldrain(url string, isFolder bool) *pixeldrainEntry {
	downloader := pixeldrainEntry{baseUrl: url, isFolder: isFolder, host: models.Host{
		Name: "Pixeldrain",
	}}
	return &downloader
}

func (dl *pixeldrainEntry) Host() *models.Host {
	return &dl.host
}

func (dl *pixeldrainEntry) BaseUrl() string {
	return dl.baseUrl
}

func (dl *pixeldrainEntry) Title() string {
	return dl.title
}

func (dl *pixeldrainEntry) ParseDownloads(c chan *[]models.Download) error {
	fileId := filepath.Base(dl.baseUrl)

	if dl.isFolder {
		res, err := models.WaitForSuccessfulRequest("https://pixeldrain.com/api/list/"+fileId, &dl.host.Timeouts)

		if err != nil {
			return err
		}

		var folder pixeldrain.FolderInfo

		err = json.NewDecoder(res.Body).Decode(&folder)

		if err != nil {
			return err
		}

		downloads := make([]models.Download, 0)

		for _, info := range folder.Files {
			savePath, err := filepath.Abs("./downloads/" + info.Name)

			if err != nil {
				panic(err)
			}

			downloads = append(downloads, models.Download{
				Url:  "https://pixeldrain.com/api/file/" + info.Id,
				Path: savePath,
				Host: &dl.host,
			})
		}

		c <- &downloads

	} else {
		res, err := models.WaitForSuccessfulRequest("https://pixeldrain.com/api/file/"+fileId+"/info", &dl.host.Timeouts)

		if err != nil {
			return err
		}

		var info pixeldrain.FileInfo

		err = json.NewDecoder(res.Body).Decode(&info)

		if err != nil {
			return err
		}

		savePath, err := filepath.Abs("./downloads/" + info.Name)

		if err != nil {
			panic(err)
		}

		downloads := make([]models.Download, 1)
		downloads[0] = models.Download{
			Url:  "https://pixeldrain.com/api/file/" + info.Id,
			Path: savePath,
			Host: &dl.host,
		}

		c <- &downloads
	}

	return nil
}
