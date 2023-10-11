package notifier

import (
	"fmt"
	"sync"
)

type ExecutionContextFunc func(*ExecutionContext)

type ExecutionContext struct {
	Parameter      any
	handlers       []ExecutionContextFunc
	executionIndex int
}

func (e *ExecutionContext) Next() {
	e.executionIndex++
	e.handlers[e.executionIndex](e)
}

type dataHandler struct {
	handlers    []ExecutionContextFunc
	middlewares []ExecutionContextFunc
}

var subscriptions map[string]dataHandler
var mutex sync.RWMutex

func init() {
	subscriptions = make(map[string]dataHandler)
	mutex = sync.RWMutex{}
}

func RegisterHandler[T any](fn func(T)) {
	key := fmt.Sprintf("%T", *new(T))
	mutex.Lock()
	defer mutex.Unlock()

	subscription := subscriptions[key]
	subscription.handlers = append(subscription.handlers, func(ctx *ExecutionContext) {
		fn(ctx.Parameter.(T))
	})
	subscriptions[key] = subscription
}

func RegisterMiddleware[T any](fn ExecutionContextFunc) {
	key := fmt.Sprintf("%T", *new(T))
	mutex.Lock()
	defer mutex.Unlock()

	subscription := subscriptions[key]
	subscription.middlewares = append(subscription.middlewares, fn)
	subscriptions[key] = subscription
}

func Publish[T any](data T) {
	key := fmt.Sprintf("%T", data)
	mutex.RLock()
	defer mutex.RUnlock()

	handlers, found := subscriptions[key]
	if !found {
		return
	}

	for _, handler := range handlers.handlers {
		if len(handlers.handlers) == 0 {
			continue
		}

		go func(hFunc ExecutionContextFunc) {
			ctx := ExecutionContext{
				executionIndex: -1,
				Parameter:      data,
				handlers:       []ExecutionContextFunc{},
			}
			ctx.handlers = append(ctx.handlers, handlers.middlewares...)
			ctx.handlers = append(ctx.handlers, hFunc) // Final function

			ctx.Next() // Execute chain
		}(handler)
	}
}
