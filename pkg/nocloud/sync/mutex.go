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
	var spm *sync.Mutex
	if m.mm[key] == nil {
		m.mm[key] = &sync.Mutex{}
	}
	spm = m.mm[key]
	m.m.Unlock()
	spm.Lock()
}

func (m *mutexM) Unlock(key string) {
	m.m.Lock()
	var spm *sync.Mutex
	if m.mm[key] == nil {
		m.mm[key] = &sync.Mutex{}
	}
	spm = m.mm[key]
	m.m.Unlock()
	spm.Unlock()
}
