package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadCronLocation(t *testing.T) {
	location, err := loadCronLocation("America/New_York")
	require.NoError(t, err)
	require.Equal(t, "America/New_York", location.String())

	_, err = loadCronLocation("not-a-timezone")
	require.ErrorContains(t, err, "invalid TIMEZONE")
}

func TestFlagHelpers(t *testing.T) {
	args := []string{"--send", "--manifest-output", "manifest.json"}
	require.True(t, hasFlag(args, "--send"))
	require.False(t, hasFlag(args, "--upload"))
	require.Equal(t, "manifest.json", flagValue(args, "--manifest-output"))
	require.Empty(t, flagValue(args, "--missing"))
}
