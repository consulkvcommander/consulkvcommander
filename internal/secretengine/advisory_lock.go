package secretengine

import "sync"

type AdvisoryLock struct {
	advisoryGuard *sync.RWMutex
	lockMap       map[string]*sync.RWMutex
}

func NewAdvisoryLock() *AdvisoryLock {
	return &AdvisoryLock{lockMap: map[string]*sync.RWMutex{}, advisoryGuard: &sync.RWMutex{}}
}

func (al *AdvisoryLock) Init(id string) {
	al.advisoryGuard.Lock()
	defer al.advisoryGuard.Unlock()
	al.lockMap[id] = &sync.RWMutex{}
}

// must be only called for ID's whose lock is already init safely in the past
func (al *AdvisoryLock) Lock(id string) {
	subLock := al.lockMap[id]
	subLock.Lock()
}

// must be only called for ID's whose lock is already init safely in the past
func (al *AdvisoryLock) Unlock(id string) {
	subLock := al.lockMap[id]
	subLock.Unlock()
}

// must be only called for ID's whose lock is already init safely in the past
func (al *AdvisoryLock) RLock(id string) {
	subLock := al.lockMap[id]
	subLock.RLock()
}

// must be only called for ID's whose lock is already init safely in the past
func (al *AdvisoryLock) RUnlock(id string) {
	subLock := al.lockMap[id]
	subLock.RUnlock()
}
