package task

import (
	"context"
)

type TaskRunOption func(o *taskRunConfig)

type HookFn func(context.Context)

type ShouldShowNotificationsFunc func(cmd *cobra.Command, cmdArgs []string) bool

type taskRunConfig struct {
	preHooks          []HookFn
	runUpdateCheck    bool
	runTrimLogs       bool
	showNotifications ShouldShowNotificationsFunc
}

func newRunConfig() *taskRunConfig {
	return &taskRunConfig{
		runUpdateCheck:    true,
		runTrimLogs:       false,
		showNotifications: shouldShowNotificationsDefault,
	}
}

func WithUpdateCheck(run bool) TaskRunOption {
	return func(o *taskRunConfig) {
		o.runUpdateCheck = run
	}
}

func WithTrimLogs(run bool) TaskRunOption {
	return func(o *taskRunConfig) {
		o.runTrimLogs = run
	}
}

func WithPreHook(f HookFn) TaskRunOption {
	return func(o *taskRunConfig) {
		o.preHooks = append(o.preHooks, f)
	}
}

func WithShowNotificationsFunc(f ShouldShowNotificationsFunc) TaskRunOption {
	return func(o *taskRunConfig) {
		o.showNotifications = f
	}
}

func shouldShowNotificationsDefault(_ *cobra.Command, _ []string) bool {
	return true
}
