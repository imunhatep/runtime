package runtime

import (
	"context"
	"errors"
	"testing"
)

func TestEnvironmentStop(t *testing.T) {
	t.Parallel()

	env := NewEnvironment(context.Background())
	waitCh := make(chan struct{})

	env.Go(func(ctx context.Context) error {
		return nil
	})
	env.Go(func(ctx context.Context) error {
		<-waitCh
		return nil
	})

	env.Stop()
	close(waitCh)
	err := env.Wait()
	if err != nil {
		t.Error(err)
	}
}

func TestEnvironmentError(t *testing.T) {
	t.Parallel()

	env := NewEnvironment(context.Background())

	testError := errors.New("test")

	stopCh := make(chan struct{})

	go func() {
		err := env.Wait()
		if err != testError {
			t.Error(`err != testError`)
		}
		close(stopCh)
	}()

	waitCh := make(chan struct{})

	env.Go(func(ctx context.Context) error {
		close(waitCh)
		return nil
	})

	<-waitCh
	env.Cancel(testError)

	<-stopCh
}

func TestEnvironmentGo(t *testing.T) {
	t.Parallel()

	env := NewEnvironment(context.Background())

	testError := errors.New("test")

	waitCh := make(chan struct{})

	env.Go(func(ctx context.Context) error {
		<-ctx.Done()
		return nil
	})

	env.Go(func(ctx context.Context) error {
		<-ctx.Done()
		// uncomment the next line delays test just 2 seconds.
		//time.Sleep(2 * time.Second)
		return nil
	})

	env.Go(func(ctx context.Context) error {
		<-waitCh
		return testError
	})

	close(waitCh)
	err := env.Wait()
	if err != testError {
		t.Error(`err != testError`)
	}
}

func TestEnvironmentID(t *testing.T) {
	t.Parallel()

	env := NewEnvironment(context.Background())

	idch := make(chan interface{}, 1)
	env.GoWithID(func(ctx context.Context) error {
		idch <- ctx.Value(RequestIDContextKey)
		return nil
	})

	env.Stop()
	err := env.Wait()
	if err != nil {
		t.Fatal(err)
	}

	id := <-idch
	if id == nil {
		t.Fatal(`id == nil`)
	}

	sid, ok := id.(string)
	if !ok {
		t.Error(`id is not a string`)
	}
	if len(sid) != 36 {
		t.Error(`len(sid) != 36`)
	}
}
