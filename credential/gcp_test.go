package credential

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGcpCredentialResolveCredentialTypes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"fake-access-token","token_type":"Bearer","expires_in":3600}`))
	}))
	defer server.Close()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	pemKey := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})

	files := map[string]map[string]string{
		"service_account": {
			"type":           "service_account",
			"project_id":     "fake-project",
			"private_key_id": "fake-key-id",
			"private_key":    string(pemKey),
			"client_email":   "fake@fake-project.iam.gserviceaccount.com",
			"token_uri":      server.URL,
		},
		"authorized_user": { //nolint:gosec // fake credentials
			"type":          "authorized_user",
			"client_id":     "fake-client-id",
			"client_secret": "fake-client-secret",
			"refresh_token": "fake-refresh-token",
			"token_uri":     server.URL,
		},
	}

	for name, content := range files {
		t.Run(name, func(t *testing.T) {
			data, err := json.Marshal(content)
			require.NoError(t, err)
			path := filepath.Join(t.TempDir(), "creds.json")
			require.NoError(t, os.WriteFile(path, data, 0600))

			cred := &GcpCredential{Credentials: &path}
			resolved, err := cred.Resolve(context.Background())
			require.NoError(t, err)

			gcpCred, ok := resolved.(*GcpCredential)
			require.True(t, ok)
			require.NotNil(t, gcpCred.AccessToken)
			require.Equal(t, "fake-access-token", *gcpCred.AccessToken)
		})
	}
}
