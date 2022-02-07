package pixeldrain

type FileInfo struct {
	Id               string `json:"id"`
	Name             string `json:"name"`
	Size             int    `json:"size"`
	Views            int    `json:"views"`
	BandwithUsed     int    `json:"bandwith_used"`
	BandwithUsedPaid int    `json:"bandwith_used_paid"`
	Downloads        int    `json:"downloads"`
	DateUploaded     string `json:"date_uploaded"`
	DateLastView     string `json:"date_last_view"`
	MimeType         string `json:"mime_type"`
}
