package downloaders

import "megazen/models"

type GenericDownloader interface {
	ParseDownloads(c chan *[]models.Download) error
}
