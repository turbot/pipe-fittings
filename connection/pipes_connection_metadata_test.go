package connection

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/utils"
	"math"
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
			got, err := m.handlePipesCredApiResponse(tt.args.jsonResponse, tt.args.target)
			if !tt.wantErr(t, err, fmt.Sprintf("handlePipesCredApiResponse(%v, %v)", tt.args.jsonResponse, tt.args.target)) {
				return
			}
			if tt.want.GetTtl() != constants.DefaultConnectionTtl {
				// adjust ttl to allow for time since test setup (rounding up to nearest second)
				setTo := tt.want.GetTtl() - int(math.Ceil(time.Since(startTime).Seconds()))
				tt.want.SetTtl(setTo)
			}

			assert.True(t, got.Equals(tt.want), fmt.Sprintf("handlePipesCredApiResponse(%v, %v) = %v, want %v", tt.args.jsonResponse, tt.args.target, got, tt.want))
		})
	}
}
