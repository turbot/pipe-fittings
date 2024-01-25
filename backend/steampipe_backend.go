package backend

import (
	"context"
	"database/sql"
	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
)

type SteampipeBackend struct {
	PostgresBackend
	connections     []string
	pluginInstances map[string]string
}

func NewSteampipeBackend(ctx context.Context, postgresBackend PostgresBackend) (*SteampipeBackend, error) {
	backend := &SteampipeBackend{
		PostgresBackend: postgresBackend,
		pluginInstances: make(map[string]string),
	}

	if err := backend.init(ctx); err != nil {
		return nil, err
	}
	return backend, nil
}

func (b *SteampipeBackend) init(ctx context.Context) error {
	db, err := b.Connect(ctx)
	if err != nil {
		return err
	}
	defer db.Close()

	// load plugin instances from steampipe_internal.steampipe_plugin
	return b.loadPluginInstances(db)

}

func (b *SteampipeBackend) loadPluginInstances(db *sql.DB) error {
	query := `SELECT plugin FROM steampipe_internal.steampipe_plugin;`

	// Execute the query
	rows, err := db.Query(query)
	if err != nil {
		return sperr.WrapWithMessage(err, "failed to read installed plugin from steampipe backend")
	}
	defer rows.Close()

	// Iterate over the results
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return sperr.WrapWithMessage(err, "failed to read installed plugin from steampipe backend")
		}
		// add the connection
		// TODO kai hack for now set version to latest
		// update steampipe to include version information in steampipe_plugin
		b.pluginInstances[name] = "latest"
	}

	// Check for errors from iterating over rows
	if err = rows.Err(); err != nil {
		return sperr.WrapWithMessage(err, "failed to read installed plugin from steampipe backend")
	}
	return nil
}
