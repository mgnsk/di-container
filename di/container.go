package di

import (
	"container/heap"
	"fmt"
	"reflect"
)

// An Item is something we manage in a priority queue.
type Item struct {
	Typ      reflect.Type
	Provider reflect.Value
	Value    interface{}
	Deps     []reflect.Type
	index    int // The index of the item in the heap.
}

func (item *Item) hasDep(typ reflect.Type) bool {
	for _, t := range item.Deps {
		if t == typ {
			return true
		}
	}
	return false
}

// A pqueue implements heap.Interface and holds Items.
type pqueue []*Item

func (pq pqueue) Len() int { return len(pq) }

func (pq pqueue) Less(i, j int) bool {
	if pq[j].hasDep(pq[i].Typ) {
		// right depends on left
		return true
	}
	if pq[i].hasDep(pq[j].Typ) {
		// left depends on right
		return false
	}
	return len(pq[i].Deps) < len(pq[j].Deps)
}

func (pq pqueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *pqueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*Item)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *pqueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

// Container is a generic dependency container.
type Container struct {
	items map[reflect.Type]*Item
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

	c.items[itemType] = item
}

// Resolve the container.
func (c *Container) Resolve() error {
	for _, item := range c.items {
		providerType := item.Provider.Type()
		// Range through provider arguments (dependencies of the node).
		for i := 0; i < providerType.NumIn(); i++ {
			if depItem, ok := c.items[providerType.In(i)]; ok {
				// An item with this type was already registered, add it as an edge.
				item.Deps = append(item.Deps, depItem.Typ)
			} else {
				return fmt.Errorf("Missing provider for type '%s'", providerType.In(i))
			}
		}
	}
	return nil
}

// Range over the container items in dependency order.
func (c *Container) Range(f func(item *Item)) {
	pq := c.heap()
	for pq.Len() > 0 {
		f(heap.Pop(pq).(*Item))
	}
}

// Build the container.
func (c *Container) Build() error {
	pq := c.heap()
	for pq.Len() > 0 {
		// Populate the dependencies (arguments) of the item provider function.
		var args []reflect.Value
		item := heap.Pop(pq).(*Item)
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
	tp := reflectItem(typ)
	item, ok := c.items[tp]
	if !ok {
		panic(fmt.Errorf("container: item with type '%T' not found", typ))
	}
	return item.Value
}

func (c *Container) heap() heap.Interface {
	pq := make(pqueue, len(c.items))
	i := 0
	for _, item := range c.items {
		pq[i] = item
		i++
	}
	heap.Init(&pq)
	return &pq
}

func reflectItem(typ interface{}) reflect.Type {
	itemValue := reflect.ValueOf(typ)
	if !itemValue.IsValid() {
		panic(fmt.Errorf("container: typ '%T' must be a valid value", typ))
	}

	tp := itemValue.Type()
	if tp.Kind() != reflect.Ptr {
		panic(fmt.Errorf("container: type '%T' must be passed as a pointer", typ))
	}

	return tp.Elem()
}
