package kite

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errBackgroundWorkerFailed = errors.New("background worker failed")

func TestApp_OnStop(t *testing.T) {
	t.Setenv("METRICS_PORT", "0")

	t.Run("runs hooks in reverse order", func(t *testing.T) {
		app := New()
		order := make([]int, 0, 2)

		app.OnStop(func(_ *Context) error {
			order = append(order, 1)
			return nil
		})
		app.OnStop(func(_ *Context) error {
			order = append(order, 2)
			return nil
		})

		err := app.runOnStopHooks(t.Context())

		require.NoError(t, err)
		assert.Equal(t, []int{2, 1}, order)
	})

	t.Run("joins returned errors and recovered panics", func(t *testing.T) {
		app := New()
		errHook := errors.New("stop hook failed")

		app.OnStop(func(_ *Context) error {
			return errHook
		})
		app.OnStop(func(_ *Context) error {
			panic("stop panic")
		})

		err := app.runOnStopHooks(t.Context())

		require.Error(t, err)
		assert.ErrorIs(t, err, errHook)
		assert.Contains(t, err.Error(), "panicked")
	})
}

func TestAppGo_Validation(t *testing.T) {
	t.Setenv("METRICS_PORT", "0")

	app := New()

	app.Go("", func(*Context) error { return nil })
	assert.Empty(t, app.backgroundWorkers)

	app.Go("worker", nil)
	assert.Empty(t, app.backgroundWorkers)

	app.Go("worker", func(*Context) error { return nil })
	require.Len(t, app.backgroundWorkers, 1)

	app.Go("worker", func(*Context) error { return nil })
	assert.Len(t, app.backgroundWorkers, 1)

	app.runtimeCtx = context.Background()
	app.Go("too-late", func(*Context) error { return nil })
	assert.Len(t, app.backgroundWorkers, 1)
}

func TestAppGo_WorkerErrorRequestsShutdown(t *testing.T) {
	t.Setenv("METRICS_PORT", "0")

	app := New()
	app.Go("failing-worker", func(*Context) error {
		return errBackgroundWorkerFailed
	})

	ctx := app.initRuntime(context.Background())
	var wg sync.WaitGroup

	app.startBackgroundWorkers(ctx, &wg)
	wg.Wait()

	require.ErrorIs(t, context.Cause(ctx), errBackgroundWorkerFailed)
	require.NoError(t, app.waitForRuntimeTasks(t.Context()))
}

func TestAppGo_WorkerHonorsCancellation(t *testing.T) {
	t.Setenv("METRICS_PORT", "0")

	app := New()
	started := make(chan struct{})
	stopped := make(chan struct{})

	app.Go("cancellable-worker", func(ctx *Context) error {
		close(started)
		<-ctx.Done()
		close(stopped)
		return ctx.Err()
	})

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	app.startBackgroundWorkers(ctx, &wg)
	<-started
	cancel()

	require.NoError(t, app.waitForRuntimeTasks(t.Context()))
	wg.Wait()

	select {
	case <-stopped:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("expected background worker to stop after context cancellation")
	}
}

func TestAppRun_WaitsForOnStopHooks(t *testing.T) {
	t.Setenv("METRICS_PORT", "0")

	app := New()
	onStopCalled := make(chan struct{})

	app.Go("failing-worker", func(*Context) error {
		return errBackgroundWorkerFailed
	})
	app.OnStop(func(_ *Context) error {
		time.Sleep(50 * time.Millisecond)
		close(onStopCalled)
		return nil
	})

	app.Run()

	select {
	case <-onStopCalled:
	case <-time.After(time.Second):
		t.Fatal("expected App.Run to wait for OnStop hook completion")
	}
}

func TestAppShutdown_CancelsBackgroundWorkers(t *testing.T) {
	t.Setenv("METRICS_PORT", "0")

	app := New()
	started := make(chan struct{})
	stopped := make(chan struct{})

	app.Go("shutdown-aware-worker", func(ctx *Context) error {
		close(started)
		<-ctx.Done()
		close(stopped)
		return ctx.Err()
	})

	runtimeCtx := app.initRuntime(context.Background())
	var wg sync.WaitGroup
	app.startBackgroundWorkers(runtimeCtx, &wg)
	<-started

	shutdownCtx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()

	require.NoError(t, app.Shutdown(shutdownCtx))
	wg.Wait()

	select {
	case <-stopped:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("expected Shutdown to cancel background worker")
	}
}
