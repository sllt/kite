# Lifecycle Hooks and Background Workers

Kite provides lifecycle APIs for work that should be managed by the application runtime itself.

This includes:

- synchronous startup hooks before traffic starts
- shutdown hooks during graceful shutdown
- long-running background workers tied to the app lifecycle

## OnStart

You can register a startup hook using the `a.OnStart()` method on your `app` instance.

## Usage

The method accepts a function with the signature:

The method accepts a function with the signature `func(ctx *kite.Context) error`.

- The `*kite.Context` passed to the hook is fully initialized and provides access to all dependency-injection-managed services (e.g., `ctx.Container.SQL`, `ctx.Container.Redis`).
- If any `OnStart` hook returns an error, the application will log the error and refuse to start.


### Example: Warming up a Cache

Here is an example of using `OnStart` to set an initial value in a Redis cache when the application starts.

```go
package main

import (
    "github.com/sllt/kite/pkg/kite"
)

func main() {
    a := kite.New()

    // Register an OnStart hook to warm up a cache.
    a.OnStart(func(ctx *kite.Context) error {
        ctx.Logger.Info("Warming up the cache...")

        // In a real app, this might come from a database or another service.
        cacheKey := "initial-data"
        cacheValue := "This is some data cached at startup."

        err := ctx.Redis.Set(ctx, cacheKey, cacheValue, 0).Err()
        if err != nil {
            ctx.Logger.Errorf("Failed to warm up cache: %v", err)
            return err // Return the error to halt startup if caching fails.
        }

        ctx.Logger.Info("Cache warmed up successfully!")

        return nil
    })

    // ... register your routes

    a.Run()
}
```

This ensures that critical startup tasks are completed successfully before the application begins accepting traffic.

## OnStop

You can register shutdown hooks using `a.OnStop()`.

- hooks run during graceful shutdown
- hooks run in **reverse registration order**
- all hooks are attempted, and Kite joins their errors

### Example

```go
app.OnStop(func(ctx *kite.Context) error {
    ctx.Logger.Info("flushing final state before shutdown")
    return nil
})
```

Use `OnStop` for:

- flushing buffers
- writing final checkpoints
- unregistering external leases
- closing app-level resources that should stop before the container is fully closed

## App.Go

`App.Go` registers a long-running background worker managed by Kite.

```go
app.Go("cache-warmer", func(ctx *kite.Context) error {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return nil
        case <-ticker.C:
            ctx.Logger.Info("refreshing cache")
        }
    }
})
```

### Behavior

- workers start with the application runtime
- the passed `*kite.Context` is canceled when shutdown begins
- returning `nil` or `context.Canceled` is treated as graceful exit
- returning any other error triggers application shutdown
- panics are recovered and logged

### Good use cases

- cache refresh loops
- outbox relays
- pollers / watchers
- heartbeat / lease renewals
- long-running sync workers
