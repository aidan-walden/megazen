package download_controller

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/panjf2000/ants/v2"
	"megazen/models"
	"megazen/models/downloaders"
	"net/http"
	"strings"
	"sync"
)

var downloads = make([]models.Download, 0)

type downloadController struct {
	pool *ants.PoolWithFunc
	wg   *sync.WaitGroup
}

func download(i interface{}, urls *[]models.Download) (err error) {
	n := i.(int32)
	fmt.Println("downloading", n)
	url := (*urls)[0]
	*urls = (*urls)[1:]
	defer func(url *models.Download) {
		err := url.DownloadFile()
		if err != nil {
			fmt.Println("download error", err)
		}
	}(&url)

	return
}

func New() *downloadController {
	defer ants.Release()

	var wg sync.WaitGroup

	pool, _ := ants.NewPoolWithFunc(4, func(i interface{}) {
		err := download(i, &downloads)
		if err != nil {
			fmt.Println("download error", err)
		}
		wg.Done()
	})

	return &downloadController{
		pool: pool,
		wg:   &wg,
	}
}

func (dlman *downloadController) SubmitDownload(c *gin.Context) {
	c.Header("Content-Type", "application/json")

	reqBody, err := c.GetRawData()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Internal Server Error: " + err.Error(),
		})
		return
	}

	urls := make([]string, 0)

	err = json.Unmarshal(reqBody, &urls)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Bad Request",
		})
		return
	}

	// Set up channel to receive URLs
	queue := make(chan *[]models.Download, len(urls))
	errors := make(chan error, len(urls))
	var wg sync.WaitGroup

	unknownUrls := make([]string, 0)
	createdDownloaders := make([]downloaders.GenericDownloader, 0)

	for _, url := range urls {

		if strings.Contains(url, "bunkr") {
			createdDownloaders = append(createdDownloaders, downloaders.NewBunkr(url))
		} else if strings.Contains(url, "gofile.io") {
			createdDownloaders = append(createdDownloaders, downloaders.NewGofile(url, "HARDCODED_TOKEN_LOLOLOLOLOL"))
		} else if strings.Contains(url, "cyberdrop.me/a/") {
			createdDownloaders = append(createdDownloaders, downloaders.NewCyberdrop(url))
		} else if strings.Contains(url, "putme.ga/album/") {
			createdDownloaders = append(createdDownloaders, downloaders.NewPutmega(url))
		} else if strings.Contains(url, "pixeldrain.com/u/") {
			createdDownloaders = append(createdDownloaders, downloaders.NewPixeldrain(url))
		} else {
			unknownUrls = append(unknownUrls, url)
		}

	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Downloads submitted",
		"unknownUrls": unknownUrls,
	})

	for _, downloader := range createdDownloaders {
		wg.Add(1)
		downloader := downloader
		go func() {
			defer wg.Done()
			err := downloader.ParseDownloads(queue)
			if err != nil {
				errors <- err
			}
		}()
	}

	// Wait for all downloads to parse on separate goroutine as to not hold up the request
	go func() {
		wg.Wait()
		close(queue)

		for url := range queue {
			downloads = append(downloads, *url...)
		}

		go dlman.ExecuteDownloads()
	}()

	return
}

func (dlman *downloadController) ExecuteDownloads() {
	dlman.wg.Wait()
	dlman.wg.Add(len(downloads))
	for i, _ := range downloads {
		fmt.Println("Submitting download")
		err := dlman.pool.Invoke(int32(i))
		if err != nil {
			panic(err)
			return
		}
	}

	// Commented out code below should be done in some sort of download manager that oversees all active downloads, because the ants pool contains all downloads. If we block waiting for the pool to complete, we have hanging goroutines.

	//dlman.wg.Wait()
	//
	//if len(downloads) == 0 {
	//  dlman.pool.Release()
	//	fmt.Println("All downloads complete")
	//}

	return
}
