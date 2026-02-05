package main

import (
	"fmt"
	"time"

	"github.com/sllt/kite/pkg/kite"
	"github.com/sllt/kite/pkg/kite/cmd/terminal"
)

func main() {
	// Create a new command-line application
	app := kite.NewCMD()

	// Add a sub-command "hello" with its handler, help and description
	app.SubCommand("hello", func(c *kite.Context) (any, error) {
		return "Hello World!", nil
	},
		kite.AddDescription("Print 'Hello World!'"),
		kite.AddHelp("hello world option"),
	)

	// Add a sub-command "params" with its handler, help and description
	app.SubCommand("params", func(c *kite.Context) (any, error) {
		return fmt.Sprintf("Hello %s!", c.Param("name")), nil
	})

	app.SubCommand("spinner", spinner)

	app.SubCommand("progress", progress)

	// Run the command-line application
	app.Run()
}

func spinner(ctx *kite.Context) (any, error) {
	// initialize the spinner
	sp := terminal.NewDotSpinner(ctx.Out)
	sp.Spin(ctx)

	defer sp.Stop()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(2 * time.Second):
	}

	return "Process Complete", nil
}

func progress(ctx *kite.Context) (any, error) {
	p, err := terminal.NewProgressBar(ctx.Out, 100)
	if err != nil {
		ctx.Warn("error initializing progress bar, err : %v", err)
	}

	for i := 1; i <= 100; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(50 * time.Millisecond):
			// do a time taking process or compute a small subset of a bigger problem,
			// this could be processing batches of a data set.

			// increment the progress to display on the progress bar.
			p.Incr(int64(1))
		}
	}

	return "Process Complete", nil
}
