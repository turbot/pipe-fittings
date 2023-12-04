package contexthelpers

import (
	"context"
	"github.com/turbot/pipe-fittings/constants"
	"log/slog"
	"os"
	"os/signal"
)

func StartCancelHandler(cancel context.CancelFunc) {
	sigIntChannel := make(chan os.Signal, 1)
	signal.Notify(sigIntChannel, os.Interrupt)
	go func() {
		<-sigIntChannel
		slog.Log(context.Background(), constants.LevelTrace, "cancel handler got SIGINT")
		// call context cancellation function
		cancel()
		// leave the channel open - any subsequent interrupts hits will be ignored
	}()
}
