package downloaders

import (
	"encoding/json"
	"megazen/models"
	"megazen/models/pixeldrain"
	"path/filepath"
)

type pixeldrainDownloader struct {
	models.Host
	baseUrl string
	title   string
}

func NewPixeldrain(url string) *pixeldrainDownloader {
	downloader := pixeldrainDownloader{baseUrl: url, Host: models.Host{
		Name: "Pixeldrain",
	}}
	return &downloader
}

func (dl *pixeldrainDownloader) ParseDownloads(c chan *[]models.Download) error {
	fileId := filepath.Base(dl.baseUrl)
	res, err := WaitForSuccessfulRequest("https://pixeldrain.com/api/file/"+fileId+"/info", &dl.Timeouts)

	if err != nil {
		return err
	}

	var info pixeldrain.FileInfo

	err = json.NewDecoder(res.Body).Decode(&info)

	if err != nil {
		return err
	}

	if err != nil {
		panic(err)
	}

	savePath, err := filepath.Abs("./downloads/" + info.Name)

	if err != nil {
		panic(err)
	}

	downloads := make([]models.Download, 1)
	downloads[0] = models.Download{
		Url:  "https://pixeldrain.com/api/file/" + fileId,
		Path: savePath,
		Host: &dl.Host,
	}

	c <- &downloads

	return nil
}
