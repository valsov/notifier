package notifier

import (
	"fmt"
	"sync"
)

// Function handling an ExecutionContext, allowing subsequent handlers execution with ExecutionContext.Next() method
type ExecutionContextFunc func(*ExecutionContext)

// Context of a notification handling execution
//
// Allows access to the parameter of the notification
type ExecutionContext struct {
	Parameter      any
	handlers       []ExecutionContextFunc
	executionIndex int
}

// Call the next handler (may be a middleware) in the execution chain
func (e *ExecutionContext) Next() {
	e.executionIndex++
	e.handlers[e.executionIndex](e)
}

// Handlers store, the map keys are type names
var subscriptions map[string][]ExecutionContextFunc
var middlewares []ExecutionContextFunc
var mutex sync.RWMutex

func init() {
	subscriptions = make(map[string][]ExecutionContextFunc)
	mutex = sync.RWMutex{}
}

// Register a function to execute when an event of the type T is published
func RegisterHandler[T any](fn func(T)) {
	key := fmt.Sprintf("%T", *new(T)) // Event key is type name
	mutex.Lock()
	defer mutex.Unlock()

	subscription := subscriptions[key]
	// Wrap function in an ExecutionContextFunc to enable chain calls with middlewares, without sacrificing simple functions
	subscription = append(subscription, func(ctx *ExecutionContext) {
		fn(ctx.Parameter.(T))
	})
	subscriptions[key] = subscription
}

// Register a middleware to be run before any notification handler is called
//
// Middlewares registration order matters
func RegisterMiddleware(fn ExecutionContextFunc) {
	mutex.Lock()
	defer mutex.Unlock()

	middlewares = append(middlewares, fn)
}

// Publish a notification to be handled by all matching registered handlers
//
// All handlers are executed in a new goroutine
func Publish[T any](data T) {
	key := fmt.Sprintf("%T", data)
	mutex.RLock()
	defer mutex.RUnlock()

	handlers, found := subscriptions[key]
	if !found {
		return
	}

	for _, handler := range handlers {
		go func(hFunc ExecutionContextFunc) {
			ctx := ExecutionContext{
				executionIndex: -1,
				Parameter:      data,
				handlers:       []ExecutionContextFunc{},
			}
			ctx.handlers = append(ctx.handlers, middlewares...)
			ctx.handlers = append(ctx.handlers, hFunc) // Final function

			ctx.Next() // Execute chain
		}(handler)
	}
}
