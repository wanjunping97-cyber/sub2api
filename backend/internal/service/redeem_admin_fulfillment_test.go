//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAdminFulfillmentBypassesLimitAndKeepsRedeemAffiliate(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	userID := int64(42)
	inviterID := int64(9001)
	code := &RedeemCode{
		ID: 102, Code: "ADMIN-SUCCESS", Type: RedeemTypeBalance, Value: 20, Status: StatusUnused,
	}
	redeemRepo := &paymentFulfillmentRedeemRepo{
		paymentOrderLifecycleRedeemRepo: paymentOrderLifecycleRedeemRepo{
			codesByCode: map[string]*RedeemCode{code.Code: code},
		},
	}
	userRepo := &mockUserRepo{getByIDUser: &User{ID: userID}}
	userRepo.updateBalanceFn = func(context.Context, int64, float64) error { return nil }
	cache := &paymentFulfillmentRedeemCacheStub{count: redeemMaxFailedAttempts}
	affiliateRepo := &paymentFulfillmentAffiliateRepoStub{
		inviteeSummary: &AffiliateSummary{
			UserID: userID, AffCode: "INVITEE", InviterID: &inviterID, CreatedAt: time.Now().Add(-time.Hour),
		},
		inviterSummary: &AffiliateSummary{
			UserID: inviterID, AffCode: "INVITER", CreatedAt: time.Now().Add(-2 * time.Hour),
		},
	}
	settingSvc := NewSettingService(&paymentFulfillmentSettingRepoStub{values: map[string]string{
		SettingKeyAffiliateEnabled:           "true",
		SettingKeyAffiliateRebateRate:        "10",
		SettingKeyAffiliateRebateFreezeHours: "0",
	}}, nil)
	affiliateSvc := NewAffiliateService(affiliateRepo, settingSvc, nil, nil)
	svc := NewRedeemService(redeemRepo, userRepo, nil, cache, nil, client, nil, affiliateSvc)

	result, err := svc.RedeemForAdminFulfillment(ctx, userID, code.Code)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Zero(t, cache.getCalls)
	require.Zero(t, cache.incrementCalls)
	require.Equal(t, 1, cache.acquireCalls)
	require.Equal(t, 1, cache.releaseCalls)
	require.Len(t, redeemRepo.useCalls, 1)
	require.Len(t, affiliateRepo.accrueCalls, 1)
	require.Equal(t, inviterID, affiliateRepo.accrueCalls[0].inviterID)
	require.Equal(t, userID, affiliateRepo.accrueCalls[0].inviteeUserID)
	require.InDelta(t, 2, affiliateRepo.accrueCalls[0].amount, 1e-8)
	require.Nil(t, affiliateRepo.accrueCalls[0].sourceOrderID)
}

func TestRedeemBalanceResetReplacesBalanceAndSkipsAffiliate(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	userID := int64(42)
	inviterID := int64(9001)
	code := &RedeemCode{
		ID: 103, Code: "RESET-100", Type: RedeemTypeBalanceReset, Value: 100, Status: StatusUnused,
	}
	redeemRepo := &paymentFulfillmentRedeemRepo{
		paymentOrderLifecycleRedeemRepo: paymentOrderLifecycleRedeemRepo{
			codesByCode: map[string]*RedeemCode{code.Code: code},
		},
	}
	userRepo := &mockUserRepo{getByIDUser: &User{ID: userID, Balance: 30}}
	userRepo.updateBalanceFn = func(context.Context, int64, float64) error {
		t.Fatal("balance_reset must replace via SetBalance, not UpdateBalance")
		return nil
	}
	userRepo.setBalanceFn = func(_ context.Context, id int64, value float64) (BalanceChange, error) {
		require.Equal(t, userID, id)
		require.InDelta(t, 100, value, 1e-8)
		return BalanceChange{Old: 30, New: 100}, nil
	}
	cache := &paymentFulfillmentRedeemCacheStub{}
	affiliateRepo := &paymentFulfillmentAffiliateRepoStub{
		inviteeSummary: &AffiliateSummary{
			UserID: userID, AffCode: "INVITEE", InviterID: &inviterID, CreatedAt: time.Now().Add(-time.Hour),
		},
		inviterSummary: &AffiliateSummary{
			UserID: inviterID, AffCode: "INVITER", CreatedAt: time.Now().Add(-2 * time.Hour),
		},
	}
	settingSvc := NewSettingService(&paymentFulfillmentSettingRepoStub{values: map[string]string{
		SettingKeyAffiliateEnabled:           "true",
		SettingKeyAffiliateRebateRate:        "10",
		SettingKeyAffiliateRebateFreezeHours: "0",
	}}, nil)
	affiliateSvc := NewAffiliateService(affiliateRepo, settingSvc, nil, nil)
	svc := NewRedeemService(redeemRepo, userRepo, nil, cache, nil, client, nil, affiliateSvc)

	result, err := svc.Redeem(ctx, userID, code.Code)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, RedeemTypeBalanceReset, result.Type)
	require.InDelta(t, 100, result.Value, 1e-8)
	require.Equal(t, 1, userRepo.setBalanceCalls)
	require.Zero(t, userRepo.updateBalanceCalls)
	require.Len(t, redeemRepo.useCalls, 1)
	require.Empty(t, affiliateRepo.accrueCalls)
}
