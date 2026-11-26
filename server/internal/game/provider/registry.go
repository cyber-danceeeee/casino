package provider

import (
	"sort"
	"strings"
	"sync"

	"casino/internal/game/domain"
)

type Registry struct {
	mu        sync.RWMutex
	providers map[string]domain.GameProvider
}

func NewRegistry() *Registry {
	return &Registry{providers: map[string]domain.GameProvider{}}
}

func (r *Registry) Register(provider domain.GameProvider) error {
	if provider == nil {
		return domain.ErrProviderNotFound
	}
	code := normalizeCode(provider.Code())
	if code == "" {
		return domain.ErrProviderNotFound
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.providers[code]; ok {
		return domain.ErrDuplicateProvider
	}
	r.providers[code] = provider
	return nil
}

func (r *Registry) Get(code string) (domain.GameProvider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, ok := r.providers[normalizeCode(code)]
	if !ok {
		return nil, domain.ErrProviderNotFound
	}
	return provider, nil
}

func (r *Registry) List() []domain.GameProvider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	codes := make([]string, 0, len(r.providers))
	for code := range r.providers {
		codes = append(codes, code)
	}
	sort.Strings(codes)

	out := make([]domain.GameProvider, 0, len(codes))
	for _, code := range codes {
		out = append(out, r.providers[code])
	}
	return out
}

type CallbackRegistry struct {
	mu       sync.RWMutex
	adapters map[string]domain.ProviderCallbackAdapter
}

func NewCallbackRegistry() *CallbackRegistry {
	return &CallbackRegistry{adapters: map[string]domain.ProviderCallbackAdapter{}}
}

func (r *CallbackRegistry) Register(adapter domain.ProviderCallbackAdapter) error {
	if adapter == nil {
		return domain.ErrCallbackNotFound
	}
	code := normalizeCode(adapter.Code())
	if code == "" {
		return domain.ErrCallbackNotFound
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.adapters[code]; ok {
		return domain.ErrDuplicateCallback
	}
	r.adapters[code] = adapter
	return nil
}

func (r *CallbackRegistry) Get(providerCode string) (domain.ProviderCallbackAdapter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	adapter, ok := r.adapters[normalizeCode(providerCode)]
	if !ok {
		return nil, domain.ErrCallbackNotFound
	}
	return adapter, nil
}

func (r *CallbackRegistry) List() []domain.ProviderCallbackAdapter {
	r.mu.RLock()
	defer r.mu.RUnlock()

	codes := make([]string, 0, len(r.adapters))
	for code := range r.adapters {
		codes = append(codes, code)
	}
	sort.Strings(codes)

	out := make([]domain.ProviderCallbackAdapter, 0, len(codes))
	for _, code := range codes {
		out = append(out, r.adapters[code])
	}
	return out
}

func normalizeCode(code string) string {
	return strings.ToLower(strings.TrimSpace(code))
}
