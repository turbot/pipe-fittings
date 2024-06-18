package backend

import (
	"context"
	"database/sql"
	"errors"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/modconfig"
	"log"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/turbot/pipe-fittings/sperr"
)

type SteampipeBackend struct {
	PostgresBackend
	// map of plugin versions, keyed by image ref
	PluginVersions map[string]*modconfig.PluginVersionString
}

func NewSteampipeBackend(ctx context.Context, postgresBackend PostgresBackend) (*SteampipeBackend, error) {
	backend := &SteampipeBackend{
		PostgresBackend: postgresBackend,
	}

	if err := backend.init(ctx); err != nil {
		return nil, err
	}
	return backend, nil
}

func (b *SteampipeBackend) Name() string {
	return constants.SteampipeBackendName
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
	query := `SELECT plugin, version FROM steampipe_internal.steampipe_plugin;`

	// Execute the query
	rows, err := db.Query(query)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "42703" {
			// version column does not exist - must be a pre-22 version of steampipe

			// just swallow error
			log.Println("[INFO] 'version' column does not exist in steampipe_internal.steampipe_plugin")
			return nil
		}

		return sperr.WrapWithMessage(err, "failed to read installed plugin from steampipe backend")
	}
	defer rows.Close()

	// create the plugin version map
	b.PluginVersions = make(map[string]*modconfig.PluginVersionString)
	// Iterate over the results
	for rows.Next() {
		var name, version string
		if err := rows.Scan(&name, &version); err != nil {
			return sperr.WrapWithMessage(err, "failed to read installed plugin from steampipe backend")
		}
		// add the plugin
		pluginVersion, err := modconfig.NewPluginVersionString(version)
		if err != nil {
			// ignore this plugin
			log.Printf("[WARN] failed to parse version for plugin '%s': %v", name, err)
		}
		b.PluginVersions[name] = pluginVersion
	}

	// Check for errors from iterating over rows
	if err = rows.Err(); err != nil {
		return sperr.WrapWithMessage(err, "failed to read installed plugin from steampipe backend")
	}
	return nil
}
