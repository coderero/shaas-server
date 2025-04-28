package collections

import "github.com/pocketbase/pocketbase/core"

type CollectionDefiner interface {
	Name() string
	Schema() *core.Collection
}
