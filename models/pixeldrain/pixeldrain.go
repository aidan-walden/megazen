package pixeldrain

type FileInfo struct {
	Id           string `json:"id"`
	Name         string `json:"name"`
	Size         int    `json:"size"`
	Views        int    `json:"views"`
	BandwithUsed int    `json:"bandwith_used"`
	MimeType     string `json:"mime_type"`
}

type FolderInfo struct {
	Id          string     `json:"id"`
	Title       string     `json:"title"`
	DateCreated string     `json:"date_created"`
	Files       []FileInfo `json:"files"`
}
