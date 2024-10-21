package connection

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
	"strconv"
	"testing"
)

func Test_buildPostgresConnectionString(t *testing.T) {
	type args struct {
		db       *string
		username *string
		host     *string
		port     *int
		password *string
		sslMode  *string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "all values are set",
			args: args{
				db:       utils.ToStringPointer("db"),
				username: utils.ToStringPointer("user"),
				host:     utils.ToStringPointer("host"),
				port:     utils.ToIntegerPointer(1234),
				password: utils.ToStringPointer("password"),
				sslMode:  utils.ToStringPointer("allow"),
			},
			want:    "postgresql://user:password@host:1234/db?sslmode=allow",
			wantErr: assert.NoError,
		},
		{
			name: "host and port",
			args: args{
				db:       utils.ToStringPointer("db"),
				username: utils.ToStringPointer("user"),
				host:     utils.ToStringPointer("host"),
				port:     utils.ToIntegerPointer(1234),
			},
			want:    "postgresql://user@host:1234/db",
			wantErr: assert.NoError,
		},
		{
			name: "password",
			args: args{
				db:       utils.ToStringPointer("db"),
				username: utils.ToStringPointer("user"),
				password: utils.ToStringPointer("password"),
			},
			want:    "postgresql://user:password@localhost:5432/db",
			wantErr: assert.NoError,
		},
		{
			name: "no host",
			args: args{
				db:       utils.ToStringPointer("db"),
				username: utils.ToStringPointer("user"),
				port:     utils.ToIntegerPointer(1234),
				password: utils.ToStringPointer("password"),
				sslMode:  utils.ToStringPointer("allow"),
			},
			want:    "postgresql://user:password@localhost:1234/db?sslmode=allow",
			wantErr: assert.NoError,
		},
		{
			name: "db and user only",
			args: args{
				db:       utils.ToStringPointer("db"),
				username: utils.ToStringPointer("user"),
			},
			want:    "postgresql://user@localhost:5432/db",
			wantErr: assert.NoError,
		},
		{
			name: "no db",
			args: args{
				username: utils.ToStringPointer("user"),
			},
			want:    "",
			wantErr: assert.Error,
		},
		{
			name: "no user",
			args: args{
				db: utils.ToStringPointer("db"),
			},
			want:    "",
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildPostgresConnectionString(tt.args.db, tt.args.username, tt.args.host, tt.args.port, tt.args.password, tt.args.sslMode)
			if !tt.wantErr(t, err, fmt.Sprintf("buildPostgresConnectionString(%v, %v, %v, %v, %v, %v)", tt.args.db, tt.args.username, tt.args.host, tt.args.port, tt.args.password, tt.args.sslMode)) {
				return
			}
			assert.Equalf(t, tt.want, got, "buildPostgresConnectionString(%v, %v, %v, %v, %v, %v)", tt.args.db, tt.args.username, tt.args.host, tt.args.port, tt.args.password, tt.args.sslMode)
		})
	}
}

func Test_postgresConnectionToEnvVarMap(t *testing.T) {
	type args struct {
		connectionString *string
		db               *string
		username         *string
		password         *string
		host             *string
		port             *int
		sslMode          *string
	}
	tests := []struct {
		name string
		args args
		want map[string]cty.Value
	}{
		{
			name: "connection string:  all values are set",
			args: args{
				connectionString: utils.ToStringPointer("postgresql://user:password@host:1234/db?sslmode=allow"),
			},
			want: map[string]cty.Value{
				"PGDATABASE": cty.StringVal("db"),
				"PGUSER":     cty.StringVal("user"),
				"PGPASSWORD": cty.StringVal("password"),
				"PGHOST":     cty.StringVal("host"),
				"PGPORT":     cty.StringVal(strconv.Itoa(1234)),
				"PGSSLMODE":  cty.StringVal("allow"),
			},
		},
		{
			name: "connection string: host and port",
			args: args{
				connectionString: utils.ToStringPointer("postgresql://user@host:1234/db"),
			},
			want: map[string]cty.Value{
				"PGDATABASE": cty.StringVal("db"),
				"PGUSER":     cty.StringVal("user"),
				"PGHOST":     cty.StringVal("host"),
				"PGPORT":     cty.StringVal(strconv.Itoa(1234)),
			},
		},
		{
			name: "connection string:  password",
			args: args{
				connectionString: utils.ToStringPointer("postgresql://user:password@localhost:5432/db"),
			},
			want: map[string]cty.Value{
				"PGDATABASE": cty.StringVal("db"),
				"PGUSER":     cty.StringVal("user"),
				"PGPASSWORD": cty.StringVal("password"),
				"PGHOST":     cty.StringVal("localhost"),
				"PGPORT":     cty.StringVal(strconv.Itoa(5432)),
			},
		},
		{
			name: "all values are set",
			args: args{
				db:       utils.ToStringPointer("db"),
				username: utils.ToStringPointer("user"),
				host:     utils.ToStringPointer("host"),
				port:     utils.ToIntegerPointer(1234),
				password: utils.ToStringPointer("password"),
				sslMode:  utils.ToStringPointer("allow"),
			},
			want: map[string]cty.Value{
				"PGDATABASE": cty.StringVal("db"),
				"PGUSER":     cty.StringVal("user"),
				"PGPASSWORD": cty.StringVal("password"),
				"PGHOST":     cty.StringVal("host"),
				"PGPORT":     cty.StringVal(strconv.Itoa(1234)),
				"PGSSLMODE":  cty.StringVal("allow"),
			},
		},
		{
			name: "host and port",
			args: args{
				db:       utils.ToStringPointer("db"),
				username: utils.ToStringPointer("user"),
				host:     utils.ToStringPointer("host"),
				port:     utils.ToIntegerPointer(1234),
			},
			want: map[string]cty.Value{
				"PGDATABASE": cty.StringVal("db"),
				"PGUSER":     cty.StringVal("user"),
				"PGHOST":     cty.StringVal("host"),
				"PGPORT":     cty.StringVal(strconv.Itoa(1234)),
			},
		},
		{
			name: "password",
			args: args{
				db:       utils.ToStringPointer("db"),
				username: utils.ToStringPointer("user"),
				password: utils.ToStringPointer("password"),
			},
			want: map[string]cty.Value{
				"PGDATABASE": cty.StringVal("db"),
				"PGUSER":     cty.StringVal("user"),
				"PGPASSWORD": cty.StringVal("password"),
			},
		},
		{
			name: "no host",
			args: args{
				db:       utils.ToStringPointer("db"),
				username: utils.ToStringPointer("user"),
				port:     utils.ToIntegerPointer(1234),
				password: utils.ToStringPointer("password"),
				sslMode:  utils.ToStringPointer("allow"),
			},
			want: map[string]cty.Value{
				"PGDATABASE": cty.StringVal("db"),
				"PGUSER":     cty.StringVal("user"),
				"PGPASSWORD": cty.StringVal("password"),
				"PGPORT":     cty.StringVal(strconv.Itoa(1234)),
				"PGSSLMODE":  cty.StringVal("allow"),
			},
		},
		{
			name: "db and user only",
			args: args{
				db:       utils.ToStringPointer("db"),
				username: utils.ToStringPointer("user"),
			},
			want: map[string]cty.Value{
				"PGDATABASE": cty.StringVal("db"),
				"PGUSER":     cty.StringVal("user"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vals := postgresConnectionToEnvVarMap(tt.args.connectionString, tt.args.db, tt.args.username, tt.args.password, tt.args.host, tt.args.port, tt.args.sslMode)
			assert.Equalf(t, tt.want, vals, "postgresConnectionToEnvVarMap(%v, %v, %v, %v, %v, %v, %v)", tt.args.connectionString, tt.args.db, tt.args.username, tt.args.password, tt.args.host, tt.args.port, tt.args.sslMode)
		})
	}
}
