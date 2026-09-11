package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNextBalanceResetAt_AddsSevenDaysUTC(t *testing.T) {
	t.Parallel()

	usedAt := time.Date(2026, time.September, 4, 15, 30, 0, 0, time.FixedZone("CST", 8*3600))
	got := NextBalanceResetAt(usedAt)
	want := time.Date(2026, time.September, 11, 7, 30, 0, 0, time.UTC)

	require.True(t, got.Location() == time.UTC)
	require.True(t, got.Equal(want), "got %s want %s", got, want)
	require.Equal(t, 7, BalanceResetCycleDays())
	require.Equal(t, 2, BalanceResetDueSoonDays())
	require.Equal(t, 7*24*time.Hour, BalanceResetCycle)
}
