package download_controller

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/panjf2000/ants/v2"
	"megazen/models"
	"megazen/models/extractors"
	"net/http"
	"strings"
	"sync"
)

type DownloadQueue struct {
	active   []*models.Download
	waiting  []*models.Download
	complete []*models.Download
}

var encounteredErrors = make([]*error, 0)

type downloadController struct {
	pool *ants.PoolWithFunc
	wg   *sync.WaitGroup
	mu   *sync.RWMutex
}

var downloadQueue = &DownloadQueue{
	active:   make([]*models.Download, 0),
	waiting:  make([]*models.Download, 0),
	complete: make([]*models.Download, 0),
}

func download(queue *DownloadQueue, mu *sync.RWMutex) (err error) {
	mu.Lock()
	url := (*queue).waiting[0]
	(*queue).waiting = (*queue).waiting[1:]
	(*queue).active = append((*queue).active, url)
	mu.Unlock()
	defer func(url *models.Download) {
		err := url.DownloadFile()
		if err != nil {
			fmt.Println("download error", err)
		}
		mu.Lock()
		queue = &DownloadQueue{
			active:   (*queue).active[1:],
			waiting:  (*queue).waiting,
			complete: append((*queue).complete, url),
		}
		mu.Unlock()
	}(url)

	return
}

func New(concurrentDownloadsCount int) *downloadController {
	defer ants.Release()

	var wg sync.WaitGroup
	var mu sync.RWMutex

	pool, _ := ants.NewPoolWithFunc(concurrentDownloadsCount, func(i interface{}) {
		err := download(downloadQueue, &mu)
		if err != nil {
			fmt.Println("download error", err)
		}
		defer wg.Done()
	})

	return &downloadController{
		pool: pool,
		wg:   &wg,
		mu:   &mu,
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
	errorsChannel := make(chan error, len(urls))
	var wg sync.WaitGroup

	unknownUrls := make([]string, 0)
	createdDownloaders := make([]models.FileHostEntry, 0)

	for _, url := range urls {

		if strings.Contains(url, "bunkr") {
			createdDownloaders = append(createdDownloaders, extractors.NewBunkr(url))
		} else if strings.Contains(url, "gofile.io/") {
			createdDownloaders = append(createdDownloaders, extractors.NewGofile(url, "AO3uS259LDIqUdRIXQZcDECeG2RxGKiX"))
		} else if strings.Contains(url, "cyberdrop.me/a/") {
			createdDownloaders = append(createdDownloaders, extractors.NewCyberdrop(url))
		} else if strings.Contains(url, "putme.ga/album/") || strings.Contains(url, "pixl.is/album/") {
			createdDownloaders = append(createdDownloaders, extractors.NewPutmega(url))
		} else if strings.Contains(url, "pixeldrain.com/") {
			createdDownloaders = append(createdDownloaders, extractors.NewPixeldrain(url, strings.Contains(url, "/l/")))
		} else if strings.Contains(url, "anonfiles.com/") {
			createdDownloaders = append(createdDownloaders, extractors.NewAnonfiles(url))
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
			errorsChannel <- err
		}()
	}

	// Wait for all downloads to parse on separate goroutine as to not hold up the request
	go func() {
		wg.Wait()
		close(queue)

		added := 0

		for url := range queue {
			dlman.mu.Lock()
			for _, download := range *url {
				downloadQueue.waiting = append(downloadQueue.waiting, &download)
				added++
			}
			dlman.mu.Unlock()
		}

		go dlman.ExecuteDownloads(added)
	}()

	go func() {
		err := <-errorsChannel
		if err != nil {
			encounteredErrors = append(encounteredErrors, &err)
			// fmt.Println("Parsing error", err)
		}
	}()

	return
}

func (dlman *downloadController) ExecuteDownloads(amountDownloads int) {
	dlman.wg.Add(amountDownloads)
	for i := 0; i < amountDownloads; i++ {
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

func (dlman *downloadController) GetActiveDownloads() *[]models.DownloadResponse {
	downloads := make([]models.DownloadResponse, 0)
	for _, download := range downloadQueue.active {
		downloads = append(downloads, models.DownloadResponse{
			Url:      download.Url,
			Path:     download.Path,
			Complete: download.IsComplete(),
			Total:    download.Total(),
			Current:  download.Current(),
			Progress: download.Progress(),
		})
	}
	return &downloads
}

func (dlman *downloadController) GetErrors() []*error {
	return encounteredErrors
}
