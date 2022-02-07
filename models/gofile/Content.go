package gofile

type Content struct {
	Status string `json:"status"`
	Data   struct {
		IsOwner            bool            `json:"isOwner"`
		ID                 string          `json:"id"`
		Type               string          `json:"type"`
		Name               string          `json:"name"`
		Childs             []string        `json:"childs"`
		ParentFolder       string          `json:"parentFolder"`
		Code               string          `json:"code"`
		CreateTime         int             `json:"createTime"`
		TotalDownloadCount int             `json:"totalDownloadCount"`
		TotalSize          int             `json:"totalSize"`
		Contents           map[string]File `json:"contents"`
	} `json:"data"`
}

type File struct {
	ID            string   `json:"id"`
	Type          string   `json:"type"`
	Name          string   `json:"name"`
	Childs        []string `json:"childs"`
	Code          string   `json:"code"`
	ParentFolder  string   `json:"parentFolder"`
	CreateTime    int      `json:"createTime"`
	DownloadCount int      `json:"downloadCount"`
	MD5           string   `json:"md5"`
	MimeType      string   `json:"mimeType"`
	Viruses       []string `json:"viruses"`
	DirectLink    string   `json:"directLink"`
	Link          string   `json:"link"`
	Thumbnail     string   `json:"thumbnail"`
}
