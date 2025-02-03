package redis

import "fmt"

type Entity int

const (
	UserEntity Entity = iota
	VerificationToken
)

type CacheKey struct {
	Entity     Entity
	Identifier string
}

var EntityToModelName = map[Entity]string{
	UserEntity:        "User",
	VerificationToken: "VerificationToken",
}

func (q CacheKey) String() string {
	return fmt.Sprintf("%s:%s", EntityToModelName[q.Entity], q.Identifier)
}

