package mega_api

import (
	"regexp"
	"strings"
)

type MegaAPI struct {
	key string
	url string
}

func New(key string) *MegaAPI {
	return &MegaAPI{key: key, url: "https://g.api.mega.co.nz"}
}

func (api *MegaAPI) Login(email string, password string) (string, error) {
	if &email == nil || &password == nil {
		return api.loginAnonymous()
	} else {
		return api.loginUser(email, password)
	}
}

func (*MegaAPI) loginUser(email string, password string) (string, error) {
	email = strings.ToLower(email)

}

func (*MegaAPI) loginAnonymous() (string, error) {

}

func (*MegaAPI) apiRequest(params map[string]string) (string, error) {

}

func parseUrl(url string) (string, error) {
	if strings.Contains(url, "/file/") {
		url = strings.Replace(url, " ", "", -1)
		reg, err := regexp.Compile("\\W\\w\\w\\w\\w\\w\\w\\w\\w\\W")
		if err != nil {
			return "", err
		}
		fileId := reg.FindString(url)
		fileId = fileId[1 : len(fileId)-1]
		idIndex := reg.FindStringIndex(url)
		key := url[idIndex[1]+1:]
		return fileId + "!" + key, nil
	} else if strings.Contains(url, "|") {
		reg, err := regexp.Compile("/#!(.*)")
		if err != nil {
			return "", err
		}
		path := reg.FindString(url)
		return path, nil
	}
	return "", nil
}
