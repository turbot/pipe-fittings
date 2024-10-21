package connection

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/utils"
	"testing"
)

func Test_buildMysqlConnectionString(t *testing.T) {
	type args struct {
		pDbName   *string
		pUserName *string
		pHost     *string
		pPort     *int
		pPassword *string
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
				pDbName:   utils.ToStringPointer("db"),
				pUserName: utils.ToStringPointer("user"),
				pHost:     utils.ToStringPointer("host"),
				pPort:     utils.ToIntegerPointer(1234),
				pPassword: utils.ToStringPointer("password"),
			},
			want:    "postgresql://user:password@host:1234/db?sslmode=allow",
			wantErr: assert.NoError,
		},
		{
			name: "no host",
			args: args{
				pDbName:   utils.ToStringPointer("db"),
				pUserName: utils.ToStringPointer("user"),
				pPort:     utils.ToIntegerPointer(1234),
				pPassword: utils.ToStringPointer("password"),
			},
			want:    "postgresql://user:password@localhost:1234/db?sslmode=allow",
			wantErr: assert.NoError,
		},
		{
			name: "db and user only",
			args: args{
				pDbName:   utils.ToStringPointer("db"),
				pUserName: utils.ToStringPointer("user"),
			},
			want:    "postgresql://user@localhost:5432/db",
			wantErr: assert.NoError,
		},
		{
			name: "no db",
			args: args{
				pUserName: utils.ToStringPointer("user"),
			},
			want:    "",
			wantErr: assert.Error,
		},
		{
			name: "no user",
			args: args{
				pDbName: utils.ToStringPointer("db"),
			},
			want:    "",
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildMysqlConnectionString(tt.args.pDbName, tt.args.pUserName, tt.args.pHost, tt.args.pPort, tt.args.pPassword)
			if !tt.wantErr(t, err, fmt.Sprintf("buildMysqlConnectionString(%v, %v, %v, %v, %v)", tt.args.pDbName, tt.args.pUserName, tt.args.pHost, tt.args.pPort, tt.args.pPassword)) {
				return
			}
			assert.Equalf(t, tt.want, got, "buildPostgresConnectionString(%v, %v, %v, %v, %v, %v)", tt.args.pDbName, tt.args.pUserName, tt.args.pHost, tt.args.pPort, tt.args.pPassword)
		})
	}
}
