package middleware

import (
	"fmt"
	"math/bits"
	"sync"
)

type Next[E any] func() E

type Middleware[T any, E any] func(ctx *T, next Next[E]) E

type Dispatcher[T any, E any] struct {
	mutex             *sync.Mutex
	middlewares       map[Bitmask][]Middleware[T, E]
	dispatchOperation Bitmask
}

const (
	OpGlobal = 1 << iota
	OperationStartIndex
)

func NewDispatcher[T any, E any](middlewares ...Middleware[T, E]) *Dispatcher[T, E] {
	return &Dispatcher[T, E]{
		mutex:             &sync.Mutex{},
		middlewares:       map[Bitmask][]Middleware[T, E]{OpGlobal: middlewares},
		dispatchOperation: OpGlobal,
	}
}

func (d *Dispatcher[T, E]) Use(args ...any) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	var operation Bitmask
	operation = OpGlobal
	for i := 0; i < len(args); i++ {
		switch arg := args[i].(type) {
		case Middleware[T, E]:
			if operation == OpGlobal {
				d.middlewares[operation] = append(d.middlewares[operation], arg)
			} else {
				for i := 0; i < bits.UintSize; i++ {
					flag := operation & (1 << uint(i))
					if operation.HasFlag(flag) {
						if _, ok := d.middlewares[flag]; !ok {
							d.middlewares[flag] = []Middleware[T, E]{}
						}
						d.middlewares[flag] = append(d.middlewares[flag], arg)
					}
				}
			}
		case Bitmask:
			operation = arg
		default:
			panic(fmt.Errorf("unknown Use argument of type %T", arg))
		}
	}

}

func (d *Dispatcher[T, E]) Dispatch(ctx *T, opts ...DispatchOption[T, E]) E {
	defer func() {
		d.mutex.Lock()
		d.dispatchOperation = OpGlobal
		d.mutex.Unlock()
	}()

	for _, opt := range opts {
		opt(d)
	}

	var middlewares []Middleware[T, E]
	d.mutex.Lock()

	if d.dispatchOperation == OpGlobal {
		middlewares = d.middlewares[d.dispatchOperation]
	} else {
		for i := 0; i < bits.UintSize; i++ {
			flag := d.dispatchOperation & (1 << uint(i))
			if d.dispatchOperation.HasFlag(flag) {
				middlewares = append(middlewares, d.middlewares[flag]...)
			}
		}
	}
	d.mutex.Unlock()
	return invokeMiddlewares(ctx, middlewares)
}

func invokeMiddlewares[T any, E interface{}](ctx *T, middlewares []Middleware[T, E]) E {
	mwLen := len(middlewares)
	if mwLen == 0 {
		var zero E
		return zero
	}

	mw := middlewares[0]
	return mw(ctx, func() E {
		if mwLen > 1 {
			return invokeMiddlewares(ctx, middlewares[1:])
		} else {
			return invokeMiddlewares(ctx, []Middleware[T, E]{})
		}
	})
}
