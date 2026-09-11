package dto

import (
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestUserFromServiceAdmin_MapsActivityTimestamps(t *testing.T) {
	t.Parallel()

	lastLoginAt := time.Date(2026, time.April, 20, 10, 0, 0, 0, time.UTC)
	lastActiveAt := lastLoginAt.Add(15 * time.Minute)
	lastUsedAt := lastLoginAt.Add(45 * time.Minute)
	nextResetAt := lastUsedAt.Add(7 * 24 * time.Hour)
	lastResetValue := 80.0

	out := UserFromServiceAdmin(&service.User{
		ID:                    42,
		Email:                 "admin@example.com",
		Username:              "admin",
		Role:                  service.RoleAdmin,
		Status:                service.StatusActive,
		LastActiveAt:          &lastActiveAt,
		LastUsedAt:            &lastUsedAt,
		NextBalanceResetAt:    &nextResetAt,
		LastBalanceResetValue: &lastResetValue,
	})

	require.NotNil(t, out)
	require.NotNil(t, out.LastActiveAt)
	require.NotNil(t, out.LastUsedAt)
	require.NotNil(t, out.NextBalanceResetAt)
	require.WithinDuration(t, lastActiveAt, *out.LastActiveAt, time.Second)
	require.WithinDuration(t, lastUsedAt, *out.LastUsedAt, time.Second)
	require.WithinDuration(t, nextResetAt, *out.NextBalanceResetAt, time.Second)
	require.NotNil(t, out.LastBalanceResetValue)
	require.Equal(t, 80.0, *out.LastBalanceResetValue)
}
