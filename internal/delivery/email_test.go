package delivery

import (
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/kextant/kextant-agent/internal/config"
	"github.com/kextant/kextant-agent/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestEmailRenderSubjectAndBuildMessage(t *testing.T) {
	delivery := NewEmailDelivery(&config.Config{
		EmailFrom:       "kextant@example.com",
		EmailRecipients: []string{"ops@example.com", "platform@example.com"},
		EmailSubject:    "Kextant {{.ClusterName}} {{.Date}}",
	}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	report := &types.Report{ClusterName: "prod", GeneratedAt: time.Date(2026, 5, 29, 0, 0, 0, 0, time.UTC)}

	subject, err := delivery.renderSubject(report)
	require.NoError(t, err)
	require.Equal(t, "Kextant prod May 29, 2026", subject)

	message := string(delivery.buildMessage(subject, "<p>hello</p>"))
	require.Contains(t, message, "From: kextant@example.com")
	require.Contains(t, message, "To: ops@example.com, platform@example.com")
	require.Contains(t, message, "Subject: Kextant prod May 29, 2026")
	require.True(t, strings.HasSuffix(message, "<p>hello</p>"))
}

func TestLoginAuth(t *testing.T) {
	auth := LoginAuth("user", "pass")
	proto, initial, err := auth.Start(nil)
	require.NoError(t, err)
	require.Equal(t, "LOGIN", proto)
	require.Empty(t, initial)

	response, err := auth.Next([]byte("Username:"), true)
	require.NoError(t, err)
	require.Equal(t, []byte("user"), response)

	response, err = auth.Next([]byte("Password:"), true)
	require.NoError(t, err)
	require.Equal(t, []byte("pass"), response)

	_, err = auth.Next([]byte("Unknown:"), true)
	require.ErrorContains(t, err, "unknown fromServer")

	response, err = auth.Next(nil, false)
	require.NoError(t, err)
	require.Nil(t, response)
}
