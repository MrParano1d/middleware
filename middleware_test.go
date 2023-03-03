package middleware_test

import (
	"sync"
	"testing"

	"github.com/mrparano1d/middleware"
	"github.com/stretchr/testify/assert"
)

type TestContext struct {
	a int
}

func CountMiddleware() middleware.Middleware[TestContext, error] {
	return func(ctx *TestContext, next middleware.Next[error]) error {
		ctx.a++
		return next()
	}
}

func TestMiddleware(t *testing.T) {
	app := middleware.NewDispatcher[TestContext, error]()
	app.Use(middleware.Middleware[TestContext, error](func(ctx *TestContext, next middleware.Next[error]) error {
		ctx.a++
		return next()
	}))
	app.Use(CountMiddleware())

	c := &TestContext{}
	if err := app.Dispatch(c); err != nil {
		t.Fatal(err)
		return
	}

	assert.Equal(t, 2, c.a)
}

const (
	OpTest middleware.Bitmask = middleware.OperationStartIndex << iota
	OpTest2
)

func TestOperationMiddleware(t *testing.T) {
	app := middleware.NewDispatcher[TestContext, error]()
	app.Use(CountMiddleware())
	app.Use(OpTest, CountMiddleware())
	app.Use(OpTest2, CountMiddleware())
	app.Use(OpTest|OpTest2, CountMiddleware())

	c := &TestContext{}
	if err := app.Dispatch(c, middleware.WithOperation[TestContext, error](middleware.OpGlobal|OpTest)); err != nil {
		t.Fatal(err)
		return
	}

	assert.Equal(t, 3, c.a)
}

func TestRaceMiddlewa(t *testing.T) {
	app := middleware.NewDispatcher[TestContext, error]()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		app.Use(CountMiddleware())
		app.Use(OpTest, CountMiddleware())
		app.Use(OpTest2, CountMiddleware())
		app.Use(OpTest|OpTest2, CountMiddleware())
		c := &TestContext{}
		if err := app.Dispatch(c, middleware.WithOperation[TestContext, error](middleware.OpGlobal|OpTest)); err != nil {
			panic(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		app.Use(CountMiddleware())
		app.Use(OpTest, CountMiddleware())
		app.Use(OpTest2, CountMiddleware())
		app.Use(OpTest|OpTest2, CountMiddleware())
		c := &TestContext{}
		if err := app.Dispatch(c, middleware.WithOperation[TestContext, error](middleware.OpGlobal|OpTest)); err != nil {
			panic(err)
		}
	}()

	wg.Wait()
}
