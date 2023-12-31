package contexthelpers

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
)

func StartCancelHandler(cancel context.CancelFunc) {
	sigIntChannel := make(chan os.Signal, 1)
	signal.Notify(sigIntChannel, os.Interrupt)
	go func() {
		<-sigIntChannel
		slog.Debug("cancel handler got SIGINT")
		// call context cancellation function
		cancel()
		// leave the channel open - any subsequent interrupts hits will be ignored
	}()
}
