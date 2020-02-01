package shorteners

import (
	"context"
	"fmt"
	"strings"
)

// Resolver resolves tries to resolve a link.
type Resolver interface {
	Resolve(context.Context, string) (string, error)
}

var resolvers map[string]Resolver

// Resolve attempts to resolve a given link using a list of Resolvers (e.g. for
// t.co).
func Resolve(ctx context.Context, link string) (string, error) {
	if link == "" {
		return "", fmt.Errorf("no link provided")
	}
	for prefix, resolver := range resolvers {
		if strings.HasPrefix(link, prefix) {
			return resolver.Resolve(ctx, link)
		}
	}
	return "", nil
}

func init() {
	resolvers = make(map[string]Resolver)
}

func registerResolver(prefix string, r Resolver) {
	resolvers[prefix] = r
}
