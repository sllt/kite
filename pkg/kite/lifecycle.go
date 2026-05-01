package kite

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

var (
	errBackgroundWorkerNameEmpty       = errors.New("background worker name cannot be empty")
	errBackgroundWorkerHandlerNil      = errors.New("background worker handler cannot be nil")
	errBackgroundWorkerDuplicateName   = errors.New("background worker name already registered")
	errBackgroundWorkerRegisterTooLate = errors.New("cannot register background worker after app has started")
	errBackgroundWorkerPanic           = errors.New("background worker panicked")
)

// BackgroundFunc is a lifecycle-managed background task.
// The passed Kite context is canceled when the application begins shutting down.
type BackgroundFunc func(ctx *Context) error

type backgroundWorker struct {
	name string
	fn   BackgroundFunc
}

// Go registers a long-running background worker managed by Kite's application lifecycle.
// Returning a non-nil error from the worker triggers application shutdown.
// Returning nil or context.Canceled is treated as a graceful exit.
func (a *App) Go(name string, fn BackgroundFunc) {
	name = strings.TrimSpace(name)

	switch {
	case name == "":
		a.Logger().Error(errBackgroundWorkerNameEmpty)
		return
	case fn == nil:
		a.Logger().Error(errBackgroundWorkerHandlerNil)
		return
	case a.runtimeCtx != nil:
		a.Logger().Errorf("%v: %s", errBackgroundWorkerRegisterTooLate, name)
		return
	case a.hasBackgroundWorker(name):
		a.Logger().Errorf("%v: %s", errBackgroundWorkerDuplicateName, name)
		return
	}

	a.backgroundWorkers = append(a.backgroundWorkers, backgroundWorker{
		name: name,
		fn:   fn,
	})
}

func (a *App) hasBackgroundWorker(name string) bool {
	for _, worker := range a.backgroundWorkers {
		if worker.name == name {
			return true
		}
	}

	return false
}

func (a *App) initRuntime(parent context.Context) context.Context {
	ctx, cancel := context.WithCancelCause(parent)
	a.runtimeCtx = ctx
	a.runtimeCancel = cancel

	return ctx
}

func (a *App) requestShutdown(cause error) {
	if a.runtimeCancel != nil {
		a.runtimeCancel(cause)
	}
}

func (a *App) startBackgroundWorkers(ctx context.Context, wg *sync.WaitGroup) {
	for _, worker := range a.backgroundWorkers {
		worker := worker
		wg.Add(1)
		a.runtimeTasks.Add(1)

		go func() {
			defer wg.Done()
			defer a.runtimeTasks.Done()

			a.runBackgroundWorker(ctx, worker)
		}()
	}
}

func (a *App) runBackgroundWorker(ctx context.Context, worker backgroundWorker) {
	c := newContext(nil, noopRequest{ctx: ctx}, a.container)
	start := time.Now()

	a.Logger().Infof("Starting background worker: %s", worker.name)

	var err error

	defer func() {
		if r := recover(); r != nil {
			panicRecovery(r, a.Logger())
			err = fmt.Errorf("worker %s: %w: %v", worker.name, errBackgroundWorkerPanic, r)
		}

		switch {
		case err == nil, errors.Is(err, context.Canceled):
			a.Logger().Infof("Background worker stopped gracefully: %s in %s", worker.name, time.Since(start))
		default:
			a.Logger().Errorf("Background worker failed: %s in %s, err: %v", worker.name, time.Since(start), err)
			a.requestShutdown(err)
		}
	}()

	err = worker.fn(c)
}

func (a *App) waitForRuntimeTasks(ctx context.Context) error {
	done := make(chan struct{})

	go func() {
		a.runtimeTasks.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (a *App) waitForShutdown(done <-chan struct{}, timeout time.Duration) {
	if done == nil {
		return
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-done:
	case <-timer.C:
		a.Logger().Warnf("Timed out waiting for shutdown to finish within %v", timeout)
	}
}
