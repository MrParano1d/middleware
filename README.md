# Generic Middlewares

This package implements middleware logic for Go, using the generics introduced in 1.18.

The generic approach makes it possible to define a separate context for the middlewares. Furthermore, with the second type parameter it is possible to define an own return-error type for the next() functions.

## Example 1 - Global Operation

If no operation is specified with .Use, all middlewares are always executed globally.

```go
  package main

  import (
    "fmt"
    "github.com/mrparano1d/middleware"
  )

  type TestContext struct {
    a int
  }

  // alternate middleware creation without converting the function type
  func CountMiddleware() middleware.Middleware[TestContext, error] {
    return func(ctx *TestContext, next middleware.Next[error]) error {
      ctx.a++
      return next()
    }
  }

  func main() {
    app := middleware.NewDispatcher[TestContext, error]()

    // the middleware function needs to be converted to the actual middleware.Middleware[T,E] generic
    app.Use(middleware.Middleware[TestContext, error](func(ctx *TestContext, next middleware.Next[error]) error {
        ctx.a++
        return next()
    }))
    app.Use(CountMiddleware())

    c := &TestContext{}
    // dispatched only with OpGlobal operation
    if err := app.Dispatch(c); err != nil {
        panic(err)
    }
		
    fmt.Println(c.a) // 1
  }
```

## Example 2 - Custom Operation

If one or more operations are passed as the first parameter, then the middleware will only be executed for these operations.

Note that if you define your own operations, they must always start with the initial value of middleware.OperationStartIndex.
Otherwise the global operation will be overwritten.

```go
  package main

  import (
    "fmt"
    "github.com/mrparano1d/middleware"
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

  const (
	OpTest middleware.Bitmask = middleware.OperationStartIndex << iota
	OpTest2
  )
	
  func main() {
    app := middleware.NewDispatcher[TestContext, error]()

    // global operation
    app.Use(CountMiddleware())
    // custom operation
    app.Use(OpTest, CountMiddleware())
    app.Use(OpTest2, CountMiddleware())
    // multiple custom operations
    app.Use(OpTest|OpTest2, CountMiddleware())

    c := &TestContext{}

    // dispatched with multiple operations
    if err := app.Dispatch(c, middleware.WithOperation[TestContext, error](middleware.OpGlobal|OpTest)); err != nil {
      panic(err)
    }
		
    fmt.Println(c.a) // 3
  }
```