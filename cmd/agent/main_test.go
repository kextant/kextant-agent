package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFlagHelpers(t *testing.T) {
	args := []string{"--send", "--manifest-output", "manifest.json"}
	require.True(t, hasFlag(args, "--send"))
	require.False(t, hasFlag(args, "--upload"))
	require.Equal(t, "manifest.json", flagValue(args, "--manifest-output"))
	require.Empty(t, flagValue(args, "--missing"))
}
