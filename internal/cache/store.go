package cache

import (
	"steam-fast-api/internal/models"
	"sync"
	"sync/atomic"
)

type Registry struct {
	Names     map[uint32]string
	Types     map[uint32]models.AppType
	Search    map[string][]uint32
	Links     sync.Map
	Discovery sync.Map
}

var Current atomic.Pointer[Registry]

func New(size int) *Registry {
	return &Registry{
		Names:  make(map[uint32]string, size),
		Types:  make(map[uint32]models.AppType, size),
		Search: make(map[string][]uint32, size/2),
	}
}

func (r *Registry) Set(id uint32, name string, appType models.AppType) {
	r.Names[id] = name
	r.Types[id] |= appType
}

func (r *Registry) Get(id uint32) (string, models.AppType, bool) {
	if name, ok := r.Names[id]; ok {
		return name, r.Types[id], true
	}
	if val, ok := r.Discovery.Load(id); ok {
		res := val.(discovered)
		return res.name, res.appType, true
	}
	return "", 0, false
}

type discovered struct {
	name    string
	appType models.AppType
}

func (r *Registry) Discover(id uint32, name string, appType models.AppType) {
	r.Discovery.Store(id, discovered{name, appType})
}
