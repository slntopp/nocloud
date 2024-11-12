package sync

import "sync"

type mutexM struct {
	m  *sync.Mutex
	mm map[string]*sync.Mutex
}

func newMutexMap() *mutexM {
	return &mutexM{
		m:  &sync.Mutex{},
		mm: make(map[string]*sync.Mutex),
	}
}

func (m *mutexM) Lock(key string) {
	m.m.Lock()
	defer m.m.Unlock()
	if m.mm[key] == nil {
		m.mm[key] = &sync.Mutex{}
	}
	m.mm[key].Lock()
}

func (m *mutexM) Unlock(key string) {
	m.m.Lock()
	defer m.m.Unlock()
	if m.mm[key] == nil {
		m.mm[key] = &sync.Mutex{}
	}
	m.mm[key].Unlock()
}
