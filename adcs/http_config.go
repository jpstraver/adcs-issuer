package adcs

import (
	"mime"
	"os"
	"strings"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	adcsHTTPTimeoutEnv      = "ADCS_HTTP_TIMEOUT"
	defaultADCSHTTPTimeout  = 60 * time.Second
	defaultTLSHandshakeWait = 10 * time.Second
	defaultIdleConnTimeout  = 90 * time.Second
)

func getADCSHTTPTimeout() time.Duration {
	logger := ctrl.Log.WithName("adcs-http-config")
	raw := strings.TrimSpace(os.Getenv(adcsHTTPTimeoutEnv))
	if raw == "" {
		return defaultADCSHTTPTimeout
	}

	timeout, err := time.ParseDuration(raw)
	if err != nil || timeout <= 0 {
		logger.Error(err, "Invalid timeout; using default", "env", adcsHTTPTimeoutEnv, "value", raw, "default", defaultADCSHTTPTimeout.String())
		return defaultADCSHTTPTimeout
	}

	return timeout
}

func normalizeContentType(value string) string {
	mediaType, _, err := mime.ParseMediaType(value)
	if err == nil {
		return strings.ToLower(strings.TrimSpace(mediaType))
	}
	return strings.ToLower(strings.TrimSpace(strings.SplitN(value, ";", 2)[0]))
}

func hasContentType(value, expected string) bool {
	return normalizeContentType(value) == strings.ToLower(expected)
}
