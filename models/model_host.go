package models

import (
	"github.com/panjf2000/ants"
	"net/http"
	"sync"
)

type Host struct {
	Name     string
	Headers  *map[string]string
	Cookies  []*http.Cookie
	Lock     sync.Mutex
	Timeouts int32
	Pool     *ants.Pool
	Wg       *sync.WaitGroup
}
