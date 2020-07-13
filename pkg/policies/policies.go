package policies

import (
	"context"
	"regexp"
	"sort"
	"sync"

	"github.com/rs/zerolog"
)

// Loader is used to load policies into a registry from various sources.
type Loader interface {
	Load(context.Context) ([]URLPolicy, error)
}

type staticLoader struct {
	policies []URLPolicy
}

func (s *staticLoader) Load(ctx context.Context) ([]URLPolicy, error) {
	return s.policies, nil
}

// StaticLoader implements the Loader interface and fills a registry with the
// given policies.
func StaticLoader(p []URLPolicy) Loader {
	return &staticLoader{policies: p}
}

// Policy is an action name that should be applied to an URL.
type Policy string

const REJECT = Policy("reject")
const APPROVE = Policy("approve")
const DEFAULT = Policy("default")

// URLPolicy is a combination of a Policy that should be applied to a given URL
// pattern.
type URLPolicy struct {
	ID         int
	URLPattern *regexp.Regexp
	Policy     Policy
	Weight     int
}

// Registry acts as container for policies.
type Registry struct {
	defaultPolicy Policy
	policies      []URLPolicy
	lock          sync.RWMutex
}

// NewRegistry creates a new registry and sets a default policy.
func NewRegistry(defaultPolicy Policy) *Registry {
	return &Registry{
		defaultPolicy: defaultPolicy,
		policies:      make([]URLPolicy, 0, 10),
	}
}

// Policies returns a copy of all the policies inside the registry.
func (r *Registry) Policies() []URLPolicy {
	res := append(make([]URLPolicy, 0, len(r.policies)), r.policies...)
	return res
}

type byWeight []URLPolicy

func (b byWeight) Len() int {
	return len(b)
}

func (b byWeight) Swap(i, j int) {
	oldi := b[i]
	b[i] = b[j]
	b[j] = oldi
}

func (b byWeight) Less(i, j int) bool {
	return b[i].Weight < b[j].Weight
}

// AddPolicy creates a new policy given the parameters and adds it to the
// registry.
func (r *Registry) AddPolicy(u string, policy Policy, weight int) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	pattern, err := regexp.Compile(u)
	if err != nil {
		return err
	}
	r.policies = append(r.policies, URLPolicy{
		URLPattern: pattern,
		Policy:     policy,
		Weight:     weight,
	})
	sort.Sort(byWeight(r.policies))
	return nil
}

// Load uses the given loader to fill the registry with policies.
func (r *Registry) Load(ctx context.Context, loader Loader) error {
	logger := zerolog.Ctx(ctx)
	result, err := loader.Load(ctx)
	if err != nil {
		return err
	}
	if result == nil {
		logger.Debug().Msg("Nil policies returned. Registry not updated.")
		return nil
	}
	logger.Debug().Msgf("Loaded %d policies.", len(result))
	sort.Sort(byWeight(result))
	r.lock.Lock()
	defer r.lock.Unlock()
	r.policies = result
	return nil
}

// DetermineForURL returns a policy that should apply to the given URL.
func (r *Registry) DetermineForURL(u string) Policy {
	r.lock.RLock()
	defer r.lock.RUnlock()
	for _, p := range r.policies {
		if p.URLPattern.MatchString(u) {
			return p.Policy
		}
	}
	return r.defaultPolicy
}
