package connection

import (
	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/utils"
	"testing"
)

func TestSteampipePgConnection_GetConnectionString(t *testing.T) {
	type fields struct {
		ConnectionImpl   ConnectionImpl
		ConnectionString *string
		DbName           *string
		UserName         *string
		Host             *string
		Port             *int
		Password         *string
		SearchPath       *string
		SslMode          *string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "connection string",
			fields: fields{
				ConnectionString: utils.ToStringPointer("postgresql://u1@localhost/db1"),
			},
			want: "postgresql://u1@localhost/db1",
		},
		{
			name: "defaults",
			fields: fields{
				ConnectionString: nil,
				UserName:         utils.ToStringPointer("user"),
				DbName:           utils.ToStringPointer("db"),
			},
			want: "postgresql://user@localhost:5432/db",
		},
		{
			name: "host and port",
			fields: fields{
				ConnectionString: nil,
				UserName:         utils.ToStringPointer("user"),
				DbName:           utils.ToStringPointer("db"),
				Host:             utils.ToStringPointer("host"),
				Port:             utils.ToIntegerPointer(1234),
			},
			want: "postgresql://user@host:1234/db",
		},
		{
			name: "password",
			fields: fields{
				ConnectionString: nil,
				UserName:         utils.ToStringPointer("user"),
				DbName:           utils.ToStringPointer("db"),
				Password:         utils.ToStringPointer("password"),
			},
			want: "postgresql://user:password@localhost:5432/db",
		},
		{
			name: "sslmode",
			fields: fields{
				ConnectionString: nil,
				UserName:         utils.ToStringPointer("user"),
				DbName:           utils.ToStringPointer("db"),
				SslMode:          utils.ToStringPointer("require"),
			},
			want: "postgresql://user@localhost:5432/db?sslmode=require",
		},
		{
			name: "all fields",
			fields: fields{
				UserName: utils.ToStringPointer("user"),
				DbName:   utils.ToStringPointer("db"),
				Host:     utils.ToStringPointer("host"),
				Port:     utils.ToIntegerPointer(1234),
				Password: utils.ToStringPointer("password"),
				SslMode:  utils.ToStringPointer("require"),
			},
			want: "postgresql://user:password@host:1234/db?sslmode=require",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &SteampipePgConnection{
				ConnectionImpl:   tt.fields.ConnectionImpl,
				ConnectionString: tt.fields.ConnectionString,
				DbName:           tt.fields.DbName,
				UserName:         tt.fields.UserName,
				Host:             tt.fields.Host,
				Port:             tt.fields.Port,
				Password:         tt.fields.Password,
				SearchPath:       tt.fields.SearchPath,
				SslMode:          tt.fields.SslMode,
			}
			assert.Equalf(t, tt.want, c.GetConnectionString(), "GetConnectionString()")
		})
	}
}
