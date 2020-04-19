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
	typ reflect.Type
	// Provider function for item type.
	provider reflect.Value
	// The built value of this item.
	Value reflect.Value
	// Graph node of the item.
	n *dag.Node
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
		panic("container: the type of the first return value of provider must be assignable to typ")
	}

	// If the function returns 2 values, the second must be an error.
	if providerType.NumOut() == 2 {
		errorInterface := reflect.TypeOf((*error)(nil)).Elem()
		if !providerType.Out(1).Implements(errorInterface) {
			panic("container: the type of the second return value of provider must be an error")
		}
	}

	item := &Item{
		typ:      itemType,
		provider: reflect.ValueOf(provider),
	}
	item.n = &dag.Node{Value: item}

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
	item, ok := c.items[reflectItem(typ)]
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
		providerType := item.provider.Type()
		// Range through provider arguments (dependencies of the node).
		for i := 0; i < providerType.NumIn(); i++ {
			if depItem, ok := c.items[providerType.In(i)]; ok {
				// An item with this type was already registered, add it as an edge.
				item.n.Edges = append(item.n.Edges, depItem.n)
			} else {
				return fmt.Errorf("Missing provider for type '%s'", providerType.In(i))
			}
		}

		c.g = append(c.g, item.n)
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
		providerType := item.provider.Type()

		for i := 0; i < providerType.NumIn(); i++ {
			// Since the graph is sorted in dependency order,
			// we know the item is built already.
			args = append(args, c.items[providerType.In(i)].Value)
		}

		result := item.provider.Call(args)
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
		panic("container: typ must be a valid value")
	}

	tp := itemValue.Type()

	// Interfaces are be passed as pointers
	if tp.Kind() == reflect.Ptr && tp.Elem().Kind() == reflect.Interface {
		return tp.Elem()
	} else if tp.Kind() != reflect.Ptr && !itemValue.IsZero() {
		panic("container: a non-pointer and non-interface typ must be a zero value")
	}

	return tp
}
