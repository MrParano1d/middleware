package middleware

type DispatchOption[T any, E any] func(dispatcher *Dispatcher[T, E])

func WithOperation[T any, E any](operation Bitmask) DispatchOption[T, E] {
	return func(dispatcher *Dispatcher[T, E]) {
		dispatcher.mutex.Lock()
		dispatcher.dispatchOperation = operation
		dispatcher.mutex.Unlock()
	}
}
