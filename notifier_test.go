package notifier

import (
	"sync"
	"testing"
)

func TestNext(t *testing.T) {
	testCases := []struct {
		handlersCount          int
		NextCallsCount         int
		expectedExecutionIndex int
	}{
		{handlersCount: 0, NextCallsCount: 1, expectedExecutionIndex: -1}, // No handlers
		{handlersCount: 2, NextCallsCount: 2, expectedExecutionIndex: 1},
		{handlersCount: 2, NextCallsCount: 3, expectedExecutionIndex: 1}, // One call too much
	}

	for _, tc := range testCases {
		handlers := make([]ExecutionContextFunc, tc.handlersCount)
		for i := 0; i < len(handlers); i++ {
			handlers[i] = func(c *ExecutionContext) {
				c.Next()
			}
		}
		ctx := ExecutionContext{
			executionIndex: -1,
			handlers:       handlers,
		}

		ctx.Next()
		if ctx.executionIndex != tc.expectedExecutionIndex {
			t.Fatalf("wrong final execution index. expected=%d, got=%d", tc.expectedExecutionIndex, ctx.executionIndex)
		}
	}
}

func TestRegisterHandler(t *testing.T) {
	err := RegisterHandler[int](nil)
	if err == nil {
		t.Fatal("expected error, got none")
	}

	err = RegisterHandler(func(p int) {})
	if err != nil {
		t.Fatalf("got error=%v", err)
	}
	err = RegisterHandler(func(p int) {})
	if err != nil {
		t.Fatalf("got error=%v", err)
	}
	err = RegisterHandler(func(p string) {})
	if err != nil {
		t.Fatalf("got error=%v", err)
	}

	var (
		handlers []ExecutionContextFunc
		found    bool
	)
	if handlers, found = subscriptions["int"]; !found {
		t.Fatal("int handlers not found")
	}
	if len(handlers) != 2 {
		t.Fatalf("wrong int handlers count. expected=%d, got=%d", 2, len(handlers))
	}

	if handlers, found = subscriptions["string"]; !found {
		t.Fatal("string handler not found")
	}
	if len(handlers) != 1 {
		t.Fatalf("wrong string handlers count. expected=%d, got=%d", 1, len(handlers))
	}
}

func TestRegisterMiddleware(t *testing.T) {
	err := RegisterMiddleware(nil)
	if err == nil {
		t.Fatal("expected error, got none")
	}

	err = RegisterMiddleware(func(e *ExecutionContext) {})
	if err != nil {
		t.Fatalf("got error=%v", err)
	}
	err = RegisterMiddleware(func(e *ExecutionContext) {})
	if err != nil {
		t.Fatalf("got error=%v", err)
	}

	if len(middlewares) != 2 {
		t.Fatalf("wrong middlewares count. expected=%d, got=%d", 2, len(middlewares))
	}
}

func TestPublish(t *testing.T) {
	wg := sync.WaitGroup{}
	var counter int // Need to instanciate the counter here for handlers and middlewares to access it

	testCases := []struct {
		handlers        []func(int)
		middlewares     []ExecutionContextFunc
		expectedCounter int
	}{
		{
			handlers: []func(int){
				func(i int) {
					counter += i
					wg.Done()
				},
				func(i int) {
					counter += i
					wg.Done()
				},
			},
			middlewares: []ExecutionContextFunc{
				func(ec *ExecutionContext) {
					counter += 1
					ec.Next()
				},
			},
			expectedCounter: 4, // 1 middleware * 2 handlers (1+1 + 1+1)
		},
		{
			handlers: []func(int){
				func(i int) {
					counter += i
					wg.Done()
				},
			},
			middlewares: []ExecutionContextFunc{
				func(ec *ExecutionContext) {
					counter = 5
					ec.Next()
				},
				func(ec *ExecutionContext) {
					// Don't call Next() to stop execution chain
					wg.Done()
				},
			},
			expectedCounter: 5, // From middleware, not incremented by handler since it isn't run
		},
	}

	for _, tc := range testCases {
		for _, m := range tc.middlewares {
			err := RegisterMiddleware(m)
			if err != nil {
				t.Fatalf("got error=%v", err)
			}
		}
		for _, h := range tc.handlers {
			err := RegisterHandler(h)
			if err != nil {
				t.Fatalf("got error=%v", err)
			}
		}

		counter = 0 // Reset
		wg.Add(len(tc.handlers))
		Publish(1)
		wg.Wait()

		if counter != tc.expectedCounter {
			t.Fatalf("wrong counter value. expected=%d, got=%d", tc.expectedCounter, counter)
		}
	}
}
