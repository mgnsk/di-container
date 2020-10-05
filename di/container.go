package di

import (
	"fmt"
	"reflect"
	"sync/atomic"

	"github.com/mgnsk/di-container/internal/dag"
)

// An Item is something we manage in a priority queue.
type Item struct {
	Value interface{}

	provider reflect.Value
	node     *dag.Node
	index    uint64
}

// Container is a generic dependency container.
type Container struct {
	items map[reflect.Type]*Item
	deps  dag.Graph
	index uint64
}

// NewContainer creates an empty container.
func NewContainer() *Container {
	return &Container{
		items: make(map[reflect.Type]*Item),
	}
}

// Register registers a provider function for type typ.
// provider must return the dependency pointed by typ as the first return type
// and possibly an error as the second return type.
func (c *Container) Register(provider interface{}) {
	providerType := reflect.TypeOf(provider)
	if providerType.Kind() != reflect.Func {
		panic("container: provider must be a function")
	}
	if providerType.NumOut() == 0 || providerType.NumOut() > 2 {
		panic("container: provider must return at least 1 value and not more than 2")
	}

	typ := providerType.Out(0)
	if _, ok := c.items[typ]; ok {
		panic(fmt.Errorf("container: item type '%T' is already registered", typ))
	}

	// If the function returns 2 values, the second must be an error.
	if providerType.NumOut() == 2 {
		errorInterface := reflect.TypeOf((*error)(nil)).Elem()
		if !providerType.Out(1).Implements(errorInterface) {
			panic(fmt.Errorf("container: the type '%s' of the second return value of provider must be an error", providerType.Out(1)))
		}
	}

	index := atomic.AddUint64(&c.index, 1)

	item := &Item{
		provider: reflect.ValueOf(provider),
		node:     &dag.Node{},
		index:    index - 1,
	}

	item.node.Value = item
	c.items[typ] = item
	c.deps = append(c.deps, item.node)
}

// Resolve the container.
func (c *Container) Resolve() error {
	for _, item := range c.items {
		providerType := item.provider.Type()
		// Range through provider arguments (dependencies of the node).
		for i := 0; i < providerType.NumIn(); i++ {
			if depItem, ok := c.items[providerType.In(i)]; ok {
				// An item with this type was already registered, add it as an edge.
				item.node.Edges = append(item.node.Edges, depItem.node)
			} else {
				return fmt.Errorf("Missing provider for type '%s'", providerType.In(i))
			}
		}
	}
	return c.deps.Resolve()
}

// Range over the container items in dependency order.
func (c *Container) Range(f func(item *Item) bool) {
	for _, item := range c.deps {
		if f(item.Value.(*Item)) == false {
			break
		}
	}
}

// Build the container.
func (c *Container) Build() error {
	for _, item := range c.deps {
		// Populate the dependencies (arguments) of the item provider function.
		var args []reflect.Value
		item := item.Value.(*Item)
		providerType := item.provider.Type()

		for i := 0; i < providerType.NumIn(); i++ {
			val := c.items[providerType.In(i)].Value
			args = append(args, reflect.ValueOf(val))
		}

		// Call the provider.
		result := item.provider.Call(args)
		if len(result) == 2 && !result[1].IsNil() {
			// We hardcoded max 2 return types for the provider.
			// The second value is the error.
			return result[1].Interface().(error)
		}
		if !result[0].IsValid() {
			panic("invalid value")
		}

		item.Value = result[0].Interface()
	}

	return nil
}

// Get returns a built dependency by type.
func (c *Container) Get(typ interface{}) interface{} {
	tp := reflectType(typ)
	item, ok := c.items[tp]
	if !ok {
		panic(fmt.Errorf("container: item with type '%T' not found", typ))
	}
	return item.Value
}

func reflectType(typ interface{}) reflect.Type {
	val := reflect.ValueOf(typ)
	tp := val.Type()

	if tp.Kind() != reflect.Ptr {
		panic(fmt.Errorf("container: type '%T' must be passed as a pointer", typ))
	}

	if !val.IsNil() {
		panic(fmt.Errorf("container: type '%T' must be nil", typ))
	}

	return tp.Elem()
}
