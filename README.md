# Runtime Library

The `runtime` library provides context-based goroutine management and signal handling for Go programs. It is designed to facilitate graceful shutdowns and handle specific signals like `SIGPIPE` when running as a systemd service.

## Features

- **Context-based Goroutine Management**: Manage goroutines with a base context that can be canceled, ensuring all goroutines are properly cleaned up.
- **Signal Handling**: Handle `SIGPIPE` signals to prevent program crashes when running as a systemd service.
- **Graceful Shutdown**: Provides mechanisms to gracefully stop and cancel running goroutines.

## Installation

To install the `runtime` library, add it to your `go.mod` file:

```sh
go get github.com/yourusername/runtime
```

## Usage
Example of using the `runtime` for gracefully shutting down goroutines with a default `Environment`

```go
package main

import (
	"context"
	"github.com/imunhatep/runtime"
	"log"
)

func main() {
	// If a goroutine started by Go returns non-nil error,
	// the framework calls env.Cancel(err) to signal other
	// goroutines to stop soon.
	runtime.Go(func(ctx context.Context) error {
        // Simulate work
        <-ctx.Done()
        log.Println("Goroutine stopped")
        return nil
	})

	// Stop declares no more Go is called.
	// This is optional if env.Cancel will be called
	// at some point (or by a signal).
	runtime.Stop()

	// Wait returns when all goroutines return.
	runtime.Wait()
}
```

### Creating an Environment
Create a new `Environment` to manage goroutines:

```go
package main

import (
    "context"
    "github.com/yourusername/runtime"
)

func main() {
    env := runtime.NewEnvironment(context.Background())
    // Use the environment to manage goroutines
}
```

### Starting Goroutines

Use the `Go` method to start a goroutine within the environment:

```go
env.Go(func(ctx context.Context) error {
    // Your goroutine logic here
    <-ctx.Done() // Watch for cancellation
    return nil
})
```

### Graceful Shutdown

To gracefully stop all goroutines, call the `Stop` or `Cancel` methods:

```go
// Stop the environment (no new goroutines will be started)
env.Stop()

// Cancel the environment with an error
env.Cancel(nil)

// Wait for all goroutines to finish
err := env.Wait()
if err != nil {
    // Handle the error
}
```


### Example with environment

Here is a complete example demonstrating the usage of the `runtime` library:

```go
package main

import (
    "context"
    "github.com/imunhatep/runtime"
    "log"
)

func main() {
    env := runtime.NewEnvironment(context.Background())

    env.Go(func(ctx context.Context) error {
        // Simulate work
        <-ctx.Done()
        log.Println("Goroutine stopped")
        return nil
    })

    // Simulate a signal to stop the environment
    env.Stop()

    // Wait for all goroutines to finish
    if err := env.Wait(); err != nil {
        log.Fatalf("Error: %v", err)
    }

    log.Println("All goroutines have finished")
}
```