package cache

import (
	"encoding/json"
	"sync"
)

type Storage struct {
	data map[string][]byte
	mx   *sync.RWMutex
}

func NewStorage() *Storage {
	return &Storage{
		data: make(map[string][]byte),
		mx:   &sync.RWMutex{},
	}
}

func (s *Storage) Get(id string) []byte {
	val, ok := s.data[id]
	if !ok {
		return nil
	}

	return val
}

func (s *Storage) Set(id string, data []byte) {
	s.mx.Lock()
	s.data[id] = data
	s.mx.Unlock()
}

func (s *Storage) SetFrom(id string, data interface{}) {
	b, _ := json.Marshal(data)
	s.mx.Lock()
	s.data[id] = b
	s.mx.Unlock()
}
