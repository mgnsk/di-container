package di

import (
	"fmt"
	"reflect"

	"github.com/mgnsk/di-container/dag"
)

// An Item is something we manage in a priority queue.
type Item struct {
	Typ      reflect.Type
	Provider reflect.Value
	Value    interface{}
	Node     *dag.Node
}

// Container is a generic dependency container.
type Container struct {
	items map[reflect.Type]*Item
	deps  dag.Graph
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
func (c *Container) Register(typ, provider interface{}) {
	itemType := reflectType(typ)
	if _, ok := c.items[itemType]; ok {
		panic(fmt.Errorf("container: item type '%T' is already registered", typ))
	}

	providerType := reflect.TypeOf(provider)
	if providerType.Kind() != reflect.Func {
		panic("container: provider must be a function")
	}
	if providerType.NumOut() == 0 || providerType.NumOut() > 2 {
		panic("container: provider must return at least 1 value and not more than 2")
	}

	if !providerType.Out(0).AssignableTo(itemType) {
		panic(fmt.Errorf(
			"container: the type '%s' of the first return value of provider must be assignable to typ '%s'",
			providerType.Out(0),
			itemType,
		))
	}

	// If the function returns 2 values, the second must be an error.
	if providerType.NumOut() == 2 {
		errorInterface := reflect.TypeOf((*error)(nil)).Elem()
		if !providerType.Out(1).Implements(errorInterface) {
			panic(fmt.Errorf("container: the type '%s' of the second return value of provider must be an error", providerType.Out(1)))
		}
	}

	item := &Item{
		Typ:      itemType,
		Provider: reflect.ValueOf(provider),
		Node:     &dag.Node{},
	}

	item.Node.Value = item
	c.items[itemType] = item
	c.deps = append(c.deps, item.Node)
}

// Resolve the container.
func (c *Container) Resolve() error {
	for _, item := range c.items {
		providerType := item.Provider.Type()
		// Range through provider arguments (dependencies of the node).
		for i := 0; i < providerType.NumIn(); i++ {
			if depItem, ok := c.items[providerType.In(i)]; ok {
				// An item with this type was already registered, add it as an edge.
				item.Node.Edges = append(item.Node.Edges, depItem.Node)
			} else {
				return fmt.Errorf("Missing provider for type '%s'", providerType.In(i))
			}
		}
	}
	return c.deps.Resolve()
}

// Range over the container items in dependency order.
func (c *Container) Range(f func(item *Item)) {
	for _, item := range c.deps {
		f(item.Value.(*Item))
	}
}

// Build the container.
func (c *Container) Build() error {
	for _, item := range c.deps {
		// Populate the dependencies (arguments) of the item provider function.
		var args []reflect.Value
		item := item.Value.(*Item)
		providerType := item.Provider.Type()

		for i := 0; i < providerType.NumIn(); i++ {
			val := c.items[providerType.In(i)].Value
			args = append(args, reflect.ValueOf(val))
		}

		// Call the provider.
		result := item.Provider.Call(args)
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
