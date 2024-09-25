package connection

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/utils"
	"io"
	"math"
	"strings"
	"testing"
	"time"
)

func TestPipesConnectionMetadata_handlePipesCredApiResponse(t *testing.T) {
	type args struct {
		jsonResponse json.RawMessage
		target       PipelingConnection
	}

	tests := []struct {
		name    string
		args    args
		want    PipelingConnection
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "TTL set from use_before",
			args: args{
				jsonResponse: json.RawMessage(fmt.Sprintf(`{
		 "config": {
		   "access_key": "ASIA4YFAKEKEYL4NOJ7FFAKE",
		   "secret_key": "HWaJE5K1tJ7ThisIsAFakeKeyrjXzw0uYWEaFAKE",
		   "session_token": "IQoJb3JpZ2luX2VjEDcaCXVzLWVhc3QtMSJIMEYCIQDIAj6iwoFeqa2shDDo40do11xNa5A7Ta4RaNIAAcZtuAIhAImglLjXIx85KpjFF2aOrlZ9MerHeWrssWzFooYXFAKE",
		   "expiration": "2024-08-15T00:05:07Z",
		   "regions": ["*"]
		 },
		 "created_at": "2024-02-01T12:00:07Z",
		 "updated_at": "2024-08-14T23:05:07Z",
		 "use_before": "%s"
		}`, time.Now().Add(5*time.Minute).Format(time.RFC3339))),
				target: &AwsConnection{},
			},
			want: &AwsConnection{
				ConnectionImpl: ConnectionImpl{
					Ttl: 5 * 60,
				},
				AccessKey:    utils.ToStringPointer("ASIA4YFAKEKEYL4NOJ7FFAKE"),
				SecretKey:    utils.ToStringPointer("HWaJE5K1tJ7ThisIsAFakeKeyrjXzw0uYWEaFAKE"),
				SessionToken: utils.ToStringPointer("IQoJb3JpZ2luX2VjEDcaCXVzLWVhc3QtMSJIMEYCIQDIAj6iwoFeqa2shDDo40do11xNa5A7Ta4RaNIAAcZtuAIhAImglLjXIx85KpjFF2aOrlZ9MerHeWrssWzFooYXFAKE"),
			},
			wantErr: assert.NoError,
		},
		{
			name: "Default TTL",
			args: args{
				jsonResponse: json.RawMessage(`{
  "config": {
    "access_key": "ASIA4YFAKEKEYL4NOJ7FFAKE",
    "secret_key": "HWaJE5K1tJ7ThisIsAFakeKeyrjXzw0uYWEaFAKE",
    "session_token": "IQoJb3JpZ2luX2VjEDcaCXVzLWVhc3QtMSJIMEYCIQDIAj6iwoFeqa2shDDo40do11xNa5A7Ta4RaNIAAcZtuAIhAImglLjXIx85KpjFF2aOrlZ9MerHeWrssWzFooYXFAKE",
    "expiration": "2024-08-15T00:05:07Z",
    "regions": ["*"]
  },
  "created_at": "2024-02-01T12:00:07Z",
  "updated_at": "2024-08-14T23:05:07Z"
}`),
				target: &AwsConnection{},
			},
			want: &AwsConnection{
				ConnectionImpl: ConnectionImpl{
					Ttl: constants.DefaultConnectionTtl,
				},
				AccessKey:    utils.ToStringPointer("ASIA4YFAKEKEYL4NOJ7FFAKE"),
				SecretKey:    utils.ToStringPointer("HWaJE5K1tJ7ThisIsAFakeKeyrjXzw0uYWEaFAKE"),
				SessionToken: utils.ToStringPointer("IQoJb3JpZ2luX2VjEDcaCXVzLWVhc3QtMSJIMEYCIQDIAj6iwoFeqa2shDDo40do11xNa5A7Ta4RaNIAAcZtuAIhAImglLjXIx85KpjFF2aOrlZ9MerHeWrssWzFooYXFAKE"),
			},
			wantErr: assert.NoError,
		},
	}
	startTime := time.Now()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := PipesConnectionMetadata{}
			// Convert string to an io.ReadCloser
			reader := io.NopCloser(strings.NewReader(string(tt.args.jsonResponse)))

			err := m.handlePipesCredApiResponse(reader, tt.args.target)
			if !tt.wantErr(t, err, fmt.Sprintf("handlePipesCredApiResponse(%v, %v)", tt.args.jsonResponse, tt.args.target)) {
				return
			}
			if tt.want.GetTtl() != constants.DefaultConnectionTtl {
				// adjust ttl to allow for time since test setup (rounding up to nearest second)
				setTo := tt.want.GetTtl() - int(math.Ceil(time.Since(startTime).Seconds()))
				tt.want.SetTtl(setTo)
			}
			got := tt.args.target
			assert.True(t, got.Equals(tt.want), fmt.Sprintf("handlePipesCredApiResponse(%v, %v) = %v, want %v", string(tt.args.jsonResponse), tt.args.target, got, tt.want))
		})
	}
}

func TestPipesConnectionMetadata_validate(t *testing.T) {
	type fields struct {
		CloudHost  *string
		User       *string
		Org        *string
		Workspace  *string
		Connection *string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Valid - user provided",
			fields: fields{
				User:       utils.ToStringPointer("user"),
				Workspace:  utils.ToStringPointer("workspace"),
				Connection: utils.ToStringPointer("connection"),
			},
			wantErr: assert.NoError,
		},
		{
			name: "Valid - org provided",
			fields: fields{
				Org:        utils.ToStringPointer("org"),
				Workspace:  utils.ToStringPointer("workspace"),
				Connection: utils.ToStringPointer("connection"),
			},
			wantErr: assert.NoError,
		},
		{
			name: "Invalid - no connection",
			fields: fields{
				User:      utils.ToStringPointer("user"),
				Workspace: utils.ToStringPointer("workspace"),
			},
			wantErr: assert.Error,
		},
		{
			name: "Invalid - no workspace",
			fields: fields{
				User:       utils.ToStringPointer("user"),
				Connection: utils.ToStringPointer("connection"),
			},
			wantErr: assert.Error,
		},
		{
			name: "Invalid - no user or org",
			fields: fields{
				Workspace:  utils.ToStringPointer("workspace"),
				Connection: utils.ToStringPointer("connection"),
			},
			wantErr: assert.Error,
		},

		{
			name: "Invalid - both user and org",
			fields: fields{
				User:       utils.ToStringPointer("user"),
				Org:        utils.ToStringPointer("org"),
				Workspace:  utils.ToStringPointer("workspace"),
				Connection: utils.ToStringPointer("connection"),
			},
			wantErr: assert.Error,
		},

		{
			name: "Valid - cloud host",
			fields: fields{
				User:       utils.ToStringPointer("user"),
				Workspace:  utils.ToStringPointer("workspace"),
				Connection: utils.ToStringPointer("connection"),
				CloudHost:  utils.ToStringPointer("foo.pipes.turbot.com"),
			},
			wantErr: assert.NoError,
		},
		{
			name: "Invalid - cloud host",
			fields: fields{
				User:       utils.ToStringPointer("user"),
				Workspace:  utils.ToStringPointer("workspace"),
				Connection: utils.ToStringPointer("connection"),
				CloudHost:  utils.ToStringPointer("foo"),
			},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := PipesConnectionMetadata{
				CloudHost:  tt.fields.CloudHost,
				User:       tt.fields.User,
				Org:        tt.fields.Org,
				Workspace:  tt.fields.Workspace,
				Connection: tt.fields.Connection,
			}
			tt.wantErr(t, m.validate(), "validate()")
		})
	}
}

func TestPipesConnectionMetadata_endpoint(t *testing.T) {
	type fields struct {
		CloudHost  *string
		User       *string
		Org        *string
		Workspace  *string
		Connection *string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "User",
			fields: fields{
				User:       utils.ToStringPointer("user1"),
				Workspace:  utils.ToStringPointer("workspace"),
				Connection: utils.ToStringPointer("connection"),
			},
			want: "https://pipes.turbot.com/api/v0/user/user1/workspace/workspace/connection/connection/private",
		},
		{
			name: "Org",
			fields: fields{
				Org:        utils.ToStringPointer("org1"),
				Workspace:  utils.ToStringPointer("workspace"),
				Connection: utils.ToStringPointer("connection"),
			},
			want: "https://pipes.turbot.com/api/v0/org/org1/workspace/workspace/connection/connection/private",
		},

		{
			name: "CloudHost and user",
			fields: fields{
				User:       utils.ToStringPointer("user1"),
				Workspace:  utils.ToStringPointer("workspace"),
				Connection: utils.ToStringPointer("connection"),
				CloudHost:  utils.ToStringPointer("foo.pipes.turbot.com"),
			},
			want: "https://foo.pipes.turbot.com/api/v0/user/user1/workspace/workspace/connection/connection/private",
		},
		{
			name: "CloudHost and org",
			fields: fields{
				Org:        utils.ToStringPointer("org1"),
				Workspace:  utils.ToStringPointer("workspace"),
				Connection: utils.ToStringPointer("connection"),
				CloudHost:  utils.ToStringPointer("foo.pipes.turbot.com"),
			},
			want: "https://foo.pipes.turbot.com/api/v0/org/org1/workspace/workspace/connection/connection/private",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := PipesConnectionMetadata{
				CloudHost:  tt.fields.CloudHost,
				User:       tt.fields.User,
				Org:        tt.fields.Org,
				Workspace:  tt.fields.Workspace,
				Connection: tt.fields.Connection,
			}
			assert.Equalf(t, tt.want, m.endpoint(), "endpoint()")
		})
	}
}
