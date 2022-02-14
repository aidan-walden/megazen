package extractors

import "megazen/models"

type Extractor struct {
	host      models.Host
	originUrl string
	title     string
	models.FileHostEntry
}
