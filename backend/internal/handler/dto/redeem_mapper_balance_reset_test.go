package dto

import (
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestRedeemCodeFromService_MapsNextBalanceResetAt(t *testing.T) {
	t.Parallel()

	usedAt := time.Date(2026, time.September, 4, 12, 0, 0, 0, time.UTC)
	out := RedeemCodeFromService(&service.RedeemCode{
		ID:     9,
		Code:   "RESET-80",
		Type:   service.RedeemTypeBalanceReset,
		Value:  80,
		Status: service.StatusUsed,
		UsedAt: &usedAt,
	})

	require.NotNil(t, out.NextBalanceResetAt)
	require.WithinDuration(t, service.NextBalanceResetAt(usedAt), *out.NextBalanceResetAt, time.Second)
}

func TestRedeemCodeFromService_UsesCustomNextResetAt(t *testing.T) {
	t.Parallel()

	usedAt := time.Date(2026, time.September, 4, 12, 0, 0, 0, time.UTC)
	next := time.Date(2026, time.September, 20, 0, 0, 0, 0, time.UTC)
	out := RedeemCodeFromService(&service.RedeemCode{
		ID:          11,
		Code:        "RESET-80",
		Type:        service.RedeemTypeBalanceReset,
		Value:       80,
		Status:      service.StatusUsed,
		UsedAt:      &usedAt,
		NextResetAt: &next,
	})

	require.NotNil(t, out.NextBalanceResetAt)
	require.True(t, out.NextBalanceResetAt.Equal(next))
}

func TestRedeemCodeFromService_OmitsNextBalanceResetForAdditiveBalance(t *testing.T) {
	t.Parallel()

	usedAt := time.Date(2026, time.September, 4, 12, 0, 0, 0, time.UTC)
	out := RedeemCodeFromService(&service.RedeemCode{
		ID:     10,
		Code:   "ADD-20",
		Type:   service.RedeemTypeBalance,
		Value:  20,
		Status: service.StatusUsed,
		UsedAt: &usedAt,
	})

	require.Nil(t, out.NextBalanceResetAt)
}
