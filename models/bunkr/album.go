package bunkr

type AlbumResponse struct {
	Message         string `json:"message"`
	Name            string `json:"name"`
	DownloadEnabled bool   `json:"downloadEnabled"`
	Files           []File `json:"files"`
}

type File struct {
	Name        string `json:"name"`
	Url         string `json:"url"`
	Thumb       string `json:"thumb"`
	ThumbSquare string `json:"thumbSquare"`
}
