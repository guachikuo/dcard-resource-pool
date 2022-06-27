# dcard-resource-pool

## Installation
```
go get github.com/guachikuo/dcard-resource-pool
```

## Get Started
### Basic Usage
Calling `New` to new a resource pool by providing `creator` and `destroyer` of the resource
```go
ctx := context.Background()

// define your custom creator and destroyer according to the resource
creator: func(context.Context) (any, error) { return *new(any), nil }
destroyer: func(context.Context, any) {}

maxIdleSize := 10
maxIdleTime := time.Duration(1 * time.Minute)

pl, err := pool(
    creator,
    destroyer,
    maxIdleSize,
    maxIdleTime,
)
if err != nil {
  // do something
  return
}
```

There are three methods that you can use `Acquire`, `Release`, `NumIdle`
```go
// get the resource from the pool
// this creates or returns a ready-to-use resource from the resource pool
r, err := pl.Acquire(ctx)
if err != nil {
  // do something
  return
}

// this releases the resource back to the resource pool
// if the pool is full, it will be destroyed
pl.Release(ctx, r)

// this returns the number of idle resources
n := pl.NumIdle()
```