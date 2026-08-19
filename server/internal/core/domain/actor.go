package domain

import "context"

// Actor represents an authenticated user identity.
type Actor struct {
	UID   string   `json:"uid"`
	Roles []string `json:"roles"`
	Name  string   `json:"name"`
}

// HasRole reports whether the actor holds the given role.
func (a Actor) HasRole(role string) bool {
	for _, r := range a.Roles {
		if r == role {
			return true
		}
	}
	return false
}

type actorKey struct{}

// WithActor injects an Actor into context.
func WithActor(ctx context.Context, actor Actor) context.Context {
	return context.WithValue(ctx, actorKey{}, actor)
}

// ActorFromContext extracts an Actor from context.
func ActorFromContext(ctx context.Context) (Actor, bool) {
	actor, ok := ctx.Value(actorKey{}).(Actor)
	return actor, ok
}

// validRoles lists the allowed user roles.
var validRoles = []string{"associate", "seller", "guide", "manager"}

// IsValidRole checks if a role string is one of the allowed roles.
func IsValidRole(role string) bool {
	for _, r := range validRoles {
		if r == role {
			return true
		}
	}
	return false
}

// IsValidRoles checks that every role is valid and the list is non-empty.
func IsValidRoles(roles []string) bool {
	if len(roles) == 0 {
		return false
	}
	for _, r := range roles {
		if !IsValidRole(r) {
			return false
		}
	}
	return true
}
