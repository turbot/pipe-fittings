package db_common

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/error_helpers"
)

type NotificationListener struct {
	notifications []*pgconn.Notification
	conn          *pgx.Conn

	onNotification func(*pgconn.Notification)
	mut            sync.Mutex
	cancel         context.CancelFunc
}

func NewNotificationListener(ctx context.Context, conn *pgx.Conn) (*NotificationListener, error) {
	if conn == nil {
		return nil, sperr.New("nil connection passed to NewNotificationListener")
	}

	listener := &NotificationListener{conn: conn}

	// tell the connection to listen to notifications
	listenSql := fmt.Sprintf("listen %s", constants.PostgresNotificationChannel)
	_, err := conn.Exec(ctx, listenSql)
	if err != nil {
		slog.Info("Error listening to notification channel", "error", err)
		conn.Close(ctx)
		return nil, err
	}

	// create cancel context to shutdown the listener
	cancelCtx, cancel := context.WithCancel(ctx)
	listener.cancel = cancel

	// start the goroutine to listen
	listener.listenToPgNotificationsAsync(cancelCtx)

	return listener, nil
}

func (c *NotificationListener) Stop(ctx context.Context) {
	c.conn.Close(ctx)
	// stop the listener goroutine
	c.cancel()
}

func (c *NotificationListener) RegisterListener(onNotification func(*pgconn.Notification)) {
	c.mut.Lock()
	defer c.mut.Unlock()

	c.onNotification = onNotification
	// send any notifications we have already collected
	for _, n := range c.notifications {
		onNotification(n)
	}
	// clear notifications
	c.notifications = nil
}

func (c *NotificationListener) listenToPgNotificationsAsync(ctx context.Context) {
	slog.Info("notificationListener listenToPgNotificationsAsync")

	go func() {
		for ctx.Err() == nil {
			slog.Info("Wait for notification")
			notification, err := c.conn.WaitForNotification(ctx)
			if err != nil && !error_helpers.IsContextCancelledError(err) {
				slog.Warn("Error waiting for notification", "error", err)
				return
			}

			if notification != nil {
				slog.Info("got notification")
				c.mut.Lock()
				// if we have a callback, call it
				if c.onNotification != nil {
					slog.Info("call notification handler")
					c.onNotification(notification)
				} else {
					// otherwise cache the notification
					slog.Info("cache notification")
					c.notifications = append(c.notifications, notification)
				}
				c.mut.Unlock()
				slog.Info("Handled notification")
			}
		}
	}()

	slog.Log(ctx, constants.LevelTrace, "InteractiveClient listenToPgNotificationsAsync DONE")
}
