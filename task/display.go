package task

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/go-kit/helpers"

	"github.com/spf13/cobra"
	"github.com/turbot/pipe-fittings/v2/error_helpers"
	"github.com/turbot/pipe-fittings/v2/filepaths"
	"github.com/turbot/pipe-fittings/v2/utils"
)

const (
	AvailableVersionsCacheStructVersion = 20230117
)

func (r *Runner) saveAvailableVersions(cli *CLIVersionCheckResponse) error {
	utils.LogTime("Runner.saveAvailableVersions start")
	defer utils.LogTime("Runner.saveAvailableVersions end")

	if cli == nil {
		// nothing to save
		return nil
	}

	notifs := &AvailableVersionCache{
		StructVersion: AvailableVersionsCacheStructVersion,
		CliCache:      cli,
	}
	// create the file - if it exists, it will be truncated by os.Create
	f, err := os.Create(filepaths.AvailableVersionsFilePath())
	if err != nil {
		return err
	}
	defer f.Close()
	encoder := json.NewEncoder(f)
	return encoder.Encode(notifs)
}

func (r *Runner) loadAvailableVersions() (*AvailableVersionCache, error) {
	utils.LogTime("Runner.getNotifications start")
	defer utils.LogTime("Runner.getNotifications end")

	// if file does not exist, return nil
	if !filehelpers.FileExists(filepaths.AvailableVersionsFilePath()) {
		slog.Debug("no available versions file found")
		return nil, nil
	}
	// TODO graza why not put this in the update check file (later)
	f, err := os.Open(filepaths.AvailableVersionsFilePath())
	if err != nil {
		return nil, err
	}
	notifications := &AvailableVersionCache{}
	decoder := json.NewDecoder(f)
	if err := decoder.Decode(notifications); err != nil {
		return nil, err
	}
	if err := error_helpers.CombineErrors(f.Close(), os.Remove(filepaths.AvailableVersionsFilePath())); err != nil {
		// if Go couldn't close the file handle, no matter - this was just good practise
		// if Go couldn't remove the notification file, it'll get truncated next time we try to write to it
		// worst case is that the notification gets shown more than once
		slog.Debug("could not close/delete notification file", "error", err)
	}
	return notifications, nil
}

// displayNotifications checks if there are any pending notifications to display
// and if so, displays them
// does nothing if the given command is a command where notifications are not displayed
func (r *Runner) displayNotifications(cmd *cobra.Command, cmdArgs []string) error {
	utils.LogTime("Runner.displayNotifications start")
	defer utils.LogTime("Runner.displayNotifications end")

	if !r.options.showNotifications(cmd, cmdArgs) {
		// do not do anything - just return
		return nil
	}

	availableVersions, err := r.loadAvailableVersions()
	if err != nil {
		return err
	}

	if helpers.IsNil(availableVersions) {
		return nil
	}

	table, err := availableVersions.asTable()
	if err != nil {
		return err
	}
	// table can be nil if there are no notifications to display
	if table == nil {
		return nil
	}

	fmt.Println() //nolint:forbidigo // acceptable
	table.Render()
	fmt.Println() //nolint:forbidigo // acceptable

	return nil
}
