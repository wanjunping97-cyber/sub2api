package service

import "time"

const (
	// BalanceResetCycle is the reminder window after a user redeems a
	// balance_reset code. Admins use the resulting next-reset time to
	// perform the next natural reset.
	BalanceResetCycle = 7 * 24 * time.Hour
	// BalanceResetDueSoonWindow is how far ahead of next_balance_reset_at
	// a user is treated as "due soon" in admin filters.
	BalanceResetDueSoonWindow = 2 * 24 * time.Hour

	BalanceResetDueOverdue = "overdue"
	BalanceResetDueSoon    = "due_soon"
)

// NextBalanceResetAt returns the admin reminder time for the next natural
// reset: 7 days after the given balance_reset redemption.
func NextBalanceResetAt(usedAt time.Time) time.Time {
	return usedAt.UTC().Add(BalanceResetCycle)
}

// LatestBalanceReset is the newest used balance_reset code for a user.
type LatestBalanceReset struct {
	UsedAt time.Time
	Value  float64
}

func BalanceResetCycleDays() int {
	return int(BalanceResetCycle / (24 * time.Hour))
}

func BalanceResetDueSoonDays() int {
	return int(BalanceResetDueSoonWindow / (24 * time.Hour))
}
