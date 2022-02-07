package models

import "sync"

type Host struct {
	Name     string
	Headers  *map[string]string
	Lock     sync.Mutex
	Timeouts int32
}
