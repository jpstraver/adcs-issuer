package issuers

import (
	"context"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestResolveAuthMode(t *testing.T) {
	tests := []struct {
		name        string
		specAuth    string
		envAuth     string
		wantAuth    string
		wantErrPart string
	}{
		{
			name:     "defaults to ntlm when unset",
			wantAuth: authModeNTLM,
		},
		{
			name:     "uses env when spec is unset",
			envAuth:  "kerberos",
			wantAuth: authModeKerberos,
		},
		{
			name:     "supports basic auth mode",
			specAuth: "basic",
			wantAuth: authModeBasic,
		},
		{
			name:     "uses spec value over env value",
			specAuth: "ntlm",
			envAuth:  "kerberos",
			wantAuth: authModeNTLM,
		},
		{
			name:     "normalizes casing and spaces",
			specAuth: "  KeRBeRoS  ",
			wantAuth: authModeKerberos,
		},
		{
			name:        "rejects invalid spec auth type",
			specAuth:    "oauth",
			wantErrPart: "unsupported auth mode",
		},
		{
			name:        "rejects invalid env auth mode",
			envAuth:     "oauth",
			wantErrPart: "unsupported auth mode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("ADCS_AUTH_MODE", tt.envAuth)
			got, err := resolveAuthMode(tt.specAuth)
			if tt.wantErrPart != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErrPart)
				}
				if !strings.Contains(err.Error(), tt.wantErrPart) {
					t.Fatalf("expected error containing %q, got %q", tt.wantErrPart, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.wantAuth {
				t.Fatalf("expected auth mode %q, got %q", tt.wantAuth, got)
			}
		})
	}
}

func TestGetUserPassword_NTLM_RealmOptional(t *testing.T) {
	f := newIssuerFactoryForTests(t, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "creds", Namespace: "ns1"},
		Data: map[string][]byte{
			"username": []byte("user1"),
			"password": []byte("pass1"),
		},
	})

	username, password, realm, err := f.getUserPassword(context.Background(), "creds", "ns1", authModeNTLM)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if username != "user1" || password != "pass1" || realm != "" {
		t.Fatalf("unexpected credentials returned: user=%q pass=%q realm=%q", username, password, realm)
	}
}

func TestGetUserPassword_Basic_RealmOptional(t *testing.T) {
	f := newIssuerFactoryForTests(t, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "creds", Namespace: "ns1"},
		Data: map[string][]byte{
			"username": []byte("user1"),
			"password": []byte("pass1"),
		},
	})

	username, password, realm, err := f.getUserPassword(context.Background(), "creds", "ns1", authModeBasic)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if username != "user1" || password != "pass1" || realm != "" {
		t.Fatalf("unexpected credentials returned: user=%q pass=%q realm=%q", username, password, realm)
	}
}

func TestGetUserPassword_Kerberos_RealmRequired(t *testing.T) {
	f := newIssuerFactoryForTests(t, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "creds", Namespace: "ns1"},
		Data: map[string][]byte{
			"username": []byte("user1"),
			"password": []byte("pass1"),
		},
	})

	_, _, _, err := f.getUserPassword(context.Background(), "creds", "ns1", authModeKerberos)
	if err == nil {
		t.Fatal("expected error for missing realm, got nil")
	}
	if !strings.Contains(err.Error(), "realm not set in secret") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetUserPassword_Kerberos_RealmPresent(t *testing.T) {
	f := newIssuerFactoryForTests(t, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "creds", Namespace: "ns1"},
		Data: map[string][]byte{
			"username": []byte("user1"),
			"password": []byte("pass1"),
			"realm":    []byte("EXAMPLE.COM"),
		},
	})

	username, password, realm, err := f.getUserPassword(context.Background(), "creds", "ns1", authModeKerberos)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if username != "user1" || password != "pass1" || realm != "EXAMPLE.COM" {
		t.Fatalf("unexpected credentials returned: user=%q pass=%q realm=%q", username, password, realm)
	}
}

func newIssuerFactoryForTests(t *testing.T, objects ...runtime.Object) *IssuerFactory {
	t.Helper()

	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add client-go scheme: %v", err)
	}

	cl := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objects...).Build()
	return &IssuerFactory{Client: cl}
}
