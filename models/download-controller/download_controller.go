package download_controller

import (
	"fmt"
	"github.com/panjf2000/ants/v2"
	"megazen/models"
	"megazen/models/extractors"
	"strings"
	"sync"
)

type DownloadQueue struct {
	processed []*models.Download
	waiting   []*models.Download
}

var encounteredErrors = make([]*error, 0)

type downloadController struct {
	pool *ants.PoolWithFunc
	wg   *sync.WaitGroup
	mu   *sync.RWMutex
}

var downloadQueue = &DownloadQueue{
	processed: make([]*models.Download, 0),
	waiting:   make([]*models.Download, 0),
}

func download(queue *DownloadQueue, mu *sync.RWMutex) (err error) {
	mu.Lock()
	url := queue.waiting[0]
	queue.waiting = queue.waiting[1:]
	queue.processed = append(queue.processed, url)
	mu.Unlock()
	defer func(url *models.Download) {
		err := url.DownloadFile()
		if err != nil {
			fmt.Println("download error", err)
		}
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

func (dlman *downloadController) SubmitDownload(urls *[]models.DownloadSubmission) *[]string {
	// Set up channel to receive URLs
	queue := make(chan *[]models.Download, len(*urls))
	errorsChannel := make(chan error, len(*urls))
	var wg sync.WaitGroup

	unknownUrls := make([]string, 0)
	createdDownloaders := make([]models.FileHostEntry, 0)

	for _, url := range *urls {

		if strings.Contains(url.Url, "bunkr") {
			createdDownloaders = append(createdDownloaders, extractors.NewBunkr(url.Url))
		} else if strings.Contains(url.Url, "gofile.io/") {
			createdDownloaders = append(createdDownloaders, extractors.NewGofile(url.Url, "AO3uS259LDIqUdRIXQZcDECeG2RxGKiX", url.Password))
		} else if strings.Contains(url.Url, "cyberdrop.me/a/") {
			createdDownloaders = append(createdDownloaders, extractors.NewCyberdrop(url.Url))
		} else if strings.Contains(url.Url, "putme.ga/album/") || strings.Contains(url.Url, "pixl.is/album/") {
			createdDownloaders = append(createdDownloaders, extractors.NewPutmega(url.Url))
		} else if strings.Contains(url.Url, "pixeldrain.com/") {
			createdDownloaders = append(createdDownloaders, extractors.NewPixeldrain(url.Url, strings.Contains(url.Url, "/l/")))
		} else if strings.Contains(url.Url, "anonfiles.com/") {
			createdDownloaders = append(createdDownloaders, extractors.NewAnonfiles(url.Url))
		} else {
			unknownUrls = append(unknownUrls, url.Url)
		}

	}

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

		dlman.mu.Lock()
		for url := range queue {
			for i := range *url {
				download := (*url)[i]
				downloadQueue.waiting = append(downloadQueue.waiting, &download)
				added++
			}
		}
		dlman.mu.Unlock()

		go dlman.ExecuteDownloads(added)
	}()

	go func() {
		err := <-errorsChannel
		if err != nil {
			encounteredErrors = append(encounteredErrors, &err)
			// fmt.Println("Parsing error", err)
		}
	}()

	return &unknownUrls
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
	for _, download := range downloadQueue.processed {
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
