package connections

import (
	"casino/internal/game/domain"
	"casino/internal/game/provider"
	"sort"
)

// Registry resolves each account from the database, so saves take effect in all
// HTTP handlers and workers without restarting or keeping stale credentials.
type Registry struct {
	*provider.Registry
	store *Store
}

func NewRegistry(store *Store) *Registry {
	return &Registry{Registry: provider.NewRegistry(), store: store}
}
func (r *Registry) Get(code string) (domain.GameProvider, error) {
	if p, err := r.Registry.Get(code); err == nil {
		return p, nil
	}
	p, found, err := r.store.Provider(code)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, domain.ErrProviderNotFound
	}
	return p, nil
}
func (r *Registry) List() []domain.GameProvider {
	byCode := map[string]domain.GameProvider{}
	for _, p := range r.store.Providers() {
		byCode[p.Code()] = p
	}
	for _, p := range r.Registry.List() {
		byCode[p.Code()] = p
	}
	codes := []string{}
	for c := range byCode {
		codes = append(codes, c)
	}
	sort.Strings(codes)
	out := []domain.GameProvider{}
	for _, c := range codes {
		out = append(out, byCode[c])
	}
	return out
}

type CallbackRegistry struct {
	*provider.CallbackRegistry
	store *Store
}

func NewCallbackRegistry(store *Store) *CallbackRegistry {
	return &CallbackRegistry{CallbackRegistry: provider.NewCallbackRegistry(), store: store}
}
func (r *CallbackRegistry) Get(code string) (domain.ProviderCallbackAdapter, error) {
	if p, err := r.CallbackRegistry.Get(code); err == nil {
		return p, nil
	}
	p, found, err := r.store.Callback(code)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, domain.ErrCallbackNotFound
	}
	return p, nil
}
func (r *CallbackRegistry) List() []domain.ProviderCallbackAdapter {
	byCode := map[string]domain.ProviderCallbackAdapter{}
	for _, p := range r.store.Callbacks() {
		byCode[p.Code()] = p
	}
	for _, p := range r.CallbackRegistry.List() {
		byCode[p.Code()] = p
	}
	codes := []string{}
	for c := range byCode {
		codes = append(codes, c)
	}
	sort.Strings(codes)
	out := []domain.ProviderCallbackAdapter{}
	for _, c := range codes {
		out = append(out, byCode[c])
	}
	return out
}
