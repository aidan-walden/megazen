package download_controller

import (
	"fmt"
	"megazen/models"
	"megazen/models/extractors"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"
)

type DownloadQueue struct {
	processed []*models.Download
	waiting   []*models.Download
}

var encounteredErrors = make([]error, 0)

var kemonoReg = regexp.MustCompile(`^(http)s?://(?:www\.)?(kemono|coomer)\.party/\w+/\w+/\w+/?`)

type downloadController struct {
	pool *ants.PoolWithFunc
	wg   *sync.WaitGroup
	mu   *sync.RWMutex
}

var downloadQueue = &DownloadQueue{
	processed: make([]*models.Download, 0),
	waiting:   make([]*models.Download, 0),
}

func downloadLoop(dl *models.Download) {
	for {
		err := dl.DownloadFile()
		if err != nil {
			if strings.Contains(err.Error(), "An existing connection was forcibly closed by the remote host") {
				time.Sleep(time.Second * 15)
				continue
			} else {
				fmt.Println(err)
				errs := dl.Errors()
				for _, err := range errs {
					fmt.Println(err)
				}
				encounteredErrors = append(encounteredErrors, errs...)
				break
			}
		}
		break
	}
}

func download(queue *DownloadQueue, mu *sync.RWMutex) {
	mu.Lock()
	dl := queue.waiting[0]

	queue.waiting = queue.waiting[1:]
	queue.processed = append(queue.processed, dl)
	mu.Unlock()
	if dl.Host.Wg != nil {
		dl.Host.Wg.Add(1)
		doDownloadLoop := func() {
			downloadLoop(dl)
			dl.Host.Wg.Done()
		}
		dl.Host.Pool.Submit(doDownloadLoop)
	} else {
		downloadLoop(dl)
	}

}

func New(concurrentDownloadsCount int) *downloadController {

	var wg sync.WaitGroup
	var mu sync.RWMutex

	pool, _ := ants.NewPoolWithFunc(concurrentDownloadsCount, func(i interface{}) {
		download(downloadQueue, &mu)
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

		url.Url = strings.TrimSpace(url.Url)

		if url.IsDirect {
			entry := extractors.NewDirect(url.Url)
			if url.Title != "" {
				entry.SetTitle(url.Title)
			}
			createdDownloaders = append(createdDownloaders, entry)
		} else {
			if strings.Contains(url.Url, "bunkr") {
				createdDownloaders = append(createdDownloaders, extractors.NewBunkr(url.Url))
			} else if strings.Contains(url.Url, "gofile.io/") {
				createdDownloaders = append(createdDownloaders, extractors.NewGofile(url.Url, "1J63qkFMeGDPcVWUj5GnZFoXf5QNHVhu", url.Password))
			} else if strings.Contains(url.Url, "cyberdrop.me/a/") {
				createdDownloaders = append(createdDownloaders, extractors.NewCyberdrop(url.Url))
			} else if strings.Contains(url.Url, "putme.ga/") || strings.Contains(url.Url, "pixl.is/") || strings.Contains(url.Url, "putmega.com/") || strings.Contains(url.Url, "jpg.church/") {
				createdDownloaders = append(createdDownloaders, extractors.NewPutmega(url.Url))
			} else if strings.Contains(url.Url, "pixeldrain.com/") {
				createdDownloaders = append(createdDownloaders, extractors.NewPixeldrain(url.Url, strings.Contains(url.Url, "/l/")))
			} else if strings.Contains(url.Url, "anonfiles.com/") {
				createdDownloaders = append(createdDownloaders, extractors.NewAnonfiles(url.Url))
			} else {
				unknownUrls = append(unknownUrls, url.Url)
			}
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
			encounteredErrors = append(encounteredErrors, err)
			// fmt.Println("Parsing error", err)
		}
	}()

	return &unknownUrls
}

func (dlman *downloadController) ExecuteDownloads(amountDownloads int) {

	dlman.wg.Add(amountDownloads)
	for i := 0; i < amountDownloads; i++ {
		err := dlman.pool.Invoke(int32(i))
		if err != nil {
			panic(err)
		}
	}

	// Commented out code below should be done in some sort of download manager that oversees all active downloads, because the ants pool contains all downloads. If we block waiting for the pool to complete, we have hanging goroutines.

	//dlman.wg.Wait()
	//
	//if len(downloads) == 0 {
	//  dlman.pool.Release()
	//	fmt.Println("All downloads complete")
	//}

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

func (dlman *downloadController) GetWaitingDownloads() *[]models.DownloadResponse {
	downloads := make([]models.DownloadResponse, 0)
	for _, download := range downloadQueue.waiting {
		downloads = append(downloads, models.DownloadResponse{
			Url:      download.Url,
			Path:     download.Path,
			Complete: false,
			Total:    download.Total(),
			Current:  0,
			Progress: 0,
		})
	}

	return &downloads
}

func (dlman *downloadController) GetErrors() []error {
	return encounteredErrors
}
