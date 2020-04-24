package di

import (
	"fmt"
	"io"
	"reflect"
	"sync"

	"github.com/mgnsk/di-container/dag"
)

// Item represents a single item which may depend on other nodes.
type Item struct {
	// Type of the item.
	Typ reflect.Type
	// Provider function for item type.
	Provider reflect.Value
	// The built value of this item.
	Value reflect.Value
	// Graph node of the item.
	Node *dag.Node
}

// Container is a generic dependency container.
type Container struct {
	items map[reflect.Type]*Item
	g     dag.Graph
	built bool
}

// NewContainer creates an empty container.
func NewContainer() *Container {
	return &Container{
		items: make(map[reflect.Type]*Item),
	}
}

// Register registers a provider function for type typ.
// Interfaces or pointers for typ must be passed with new(T).
// Other types must be passed as the zero value.
// provider must return the dependency type as the first return type
// and possibly an error as the second return type.
func (c *Container) Register(typ, provider interface{}) {
	itemType := reflectItem(typ)
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
	}
	item.Node = &dag.Node{Value: item}

	c.items[itemType] = item
}

// Resolve the container.
func (c *Container) Resolve() error {
	if err := c.resolve(); err != nil {
		return err
	}
	return nil
}

// Range over the container items in dependency order.
func (c *Container) Range(f func(item *Item)) {
	for _, n := range c.g {
		f(n.Value.(*Item))
	}
}

// Build the container.
func (c *Container) Build() error {
	if err := c.build(); err != nil {
		return err
	}
	return nil
}

// Get returns a built dependency by type.
func (c *Container) Get(typ interface{}) interface{} {
	tp := reflectItem(typ)
	item, ok := c.items[tp]
	if !ok {
		panic(fmt.Errorf("container: item with type '%T' not found", typ))
	}
	return item.Value.Interface()
}

// Close the container in graph order. If any item implements io.Closer, it will be closed.
func (c *Container) Close() <-chan error {
	var wg sync.WaitGroup
	errs := make(chan error)
	c.Range(func(item *Item) {
		if closer, ok := item.Value.Interface().(io.Closer); ok {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := closer.Close(); err != nil {
					errs <- err
				}
			}()
		}
	})
	go func() {
		wg.Wait()
		close(errs)
	}()

	return errs
}

func (c *Container) resolve() error {
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

		c.g = append(c.g, item.Node)
	}

	// Sort the dependency graph.
	if err := c.g.Resolve(); err != nil {
		// Empty graph or cycle detected.
		return err
	}

	return nil
}

func (c *Container) build() error {
	for _, n := range c.g {
		// Populate the dependencies (arguments) of the provider function.
		var args []reflect.Value
		item := n.Value.(*Item)
		providerType := item.Provider.Type()

		for i := 0; i < providerType.NumIn(); i++ {
			// Since the graph is sorted in dependency order,
			// we know the item is built already.
			args = append(args, c.items[providerType.In(i)].Value)
		}

		result := item.Provider.Call(args)
		if len(result) == 2 && !result[1].IsNil() {
			// We hardcoded max 2 return types for the provider.
			// The second value is the error.
			return result[1].Interface().(error)
		}

		item.Value = result[0]
	}

	return nil
}

func reflectItem(typ interface{}) reflect.Type {
	itemValue := reflect.ValueOf(typ)
	if !itemValue.IsValid() {
		panic(fmt.Errorf("container: typ '%T' must be a valid value", typ))
	}

	tp := itemValue.Type()
	if tp.Kind() != reflect.Ptr {
		panic(fmt.Errorf("container: specify type '%T' as new(T)", typ))
	}

	return tp.Elem()
}
