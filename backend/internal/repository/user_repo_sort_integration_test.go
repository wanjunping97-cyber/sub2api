//go:build integration

package repository

import (
	"fmt"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

func (s *UserRepoSuite) mustInsertUsageLog(userID int64, createdAt time.Time) {
	s.T().Helper()

	account := mustCreateAccount(s.T(), s.client, &service.Account{Name: "usage-log-account"})
	apiKey := mustCreateApiKey(s.T(), s.client, &service.APIKey{UserID: userID})

	_, err := integrationDB.ExecContext(
		s.ctx,
		`INSERT INTO usage_logs (user_id, api_key_id, account_id, model, input_tokens, output_tokens, total_cost, actual_cost, created_at)
		 VALUES ($1, $2, $3, 'gpt-test', 1, 1, 0.01, 0.01, $4)`,
		userID,
		apiKey.ID,
		account.ID,
		createdAt.UTC(),
	)
	s.Require().NoError(err)
}

func (s *UserRepoSuite) TestListWithFilters_SortByEmailAsc() {
	s.mustCreateUser(&service.User{Email: "z-last@example.com", Username: "z-user"})
	s.mustCreateUser(&service.User{Email: "a-first@example.com", Username: "a-user"})

	users, _, err := s.repo.ListWithFilters(s.ctx, pagination.PaginationParams{
		Page:      1,
		PageSize:  10,
		SortBy:    "email",
		SortOrder: "asc",
	}, service.UserListFilters{})
	s.Require().NoError(err)
	s.Require().Len(users, 2)
	s.Require().Equal("a-first@example.com", users[0].Email)
	s.Require().Equal("z-last@example.com", users[1].Email)
}

func (s *UserRepoSuite) TestList_DefaultSortByNewestFirst() {
	first := s.mustCreateUser(&service.User{Email: "first@example.com"})
	second := s.mustCreateUser(&service.User{Email: "second@example.com"})

	users, _, err := s.repo.List(s.ctx, pagination.PaginationParams{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.Require().Len(users, 2)
	s.Require().Equal(second.ID, users[0].ID)
	s.Require().Equal(first.ID, users[1].ID)
}

func (s *UserRepoSuite) TestCreateAndRead_PreservesSignupSourceAndActivityTimestamps() {
	lastLoginAt := time.Now().Add(-2 * time.Hour).UTC().Truncate(time.Microsecond)
	lastActiveAt := time.Now().Add(-30 * time.Minute).UTC().Truncate(time.Microsecond)

	created := s.mustCreateUser(&service.User{
		Email:        "identity-meta@example.com",
		SignupSource: "linuxdo",
		LastLoginAt:  &lastLoginAt,
		LastActiveAt: &lastActiveAt,
	})

	got, err := s.repo.GetByID(s.ctx, created.ID)
	s.Require().NoError(err)
	s.Require().Equal("linuxdo", got.SignupSource)
	s.Require().NotNil(got.LastLoginAt)
	s.Require().NotNil(got.LastActiveAt)
	s.Require().True(got.LastLoginAt.Equal(lastLoginAt))
	s.Require().True(got.LastActiveAt.Equal(lastActiveAt))
}

func (s *UserRepoSuite) TestUpdate_PersistsSignupSourceAndActivityTimestamps() {
	created := s.mustCreateUser(&service.User{Email: "identity-update@example.com"})
	lastLoginAt := time.Now().Add(-90 * time.Minute).UTC().Truncate(time.Microsecond)
	lastActiveAt := time.Now().Add(-15 * time.Minute).UTC().Truncate(time.Microsecond)

	created.SignupSource = "oidc"
	created.LastLoginAt = &lastLoginAt
	created.LastActiveAt = &lastActiveAt

	s.Require().NoError(s.repo.Update(s.ctx, created, service.UserUpdateFields{SignupSource: true, LastLoginAt: true, LastActiveAt: true}))

	got, err := s.repo.GetByID(s.ctx, created.ID)
	s.Require().NoError(err)
	s.Require().Equal("oidc", got.SignupSource)
	s.Require().NotNil(got.LastLoginAt)
	s.Require().NotNil(got.LastActiveAt)
	s.Require().True(got.LastLoginAt.Equal(lastLoginAt))
	s.Require().True(got.LastActiveAt.Equal(lastActiveAt))
}

func (s *UserRepoSuite) TestListWithFilters_SortByLastActiveAtAsc() {
	earlier := time.Now().Add(-3 * time.Hour).UTC().Truncate(time.Microsecond)
	later := time.Now().Add(-45 * time.Minute).UTC().Truncate(time.Microsecond)

	s.mustCreateUser(&service.User{Email: "nil-active@example.com"})
	s.mustCreateUser(&service.User{Email: "later-active@example.com", LastActiveAt: &later})
	s.mustCreateUser(&service.User{Email: "earlier-active@example.com", LastActiveAt: &earlier})

	users, _, err := s.repo.ListWithFilters(s.ctx, pagination.PaginationParams{
		Page:      1,
		PageSize:  10,
		SortBy:    "last_active_at",
		SortOrder: "asc",
	}, service.UserListFilters{})
	s.Require().NoError(err)
	s.Require().Len(users, 3)
	s.Require().Equal("earlier-active@example.com", users[0].Email)
	s.Require().Equal("later-active@example.com", users[1].Email)
	s.Require().Equal("nil-active@example.com", users[2].Email)
}

func (s *UserRepoSuite) TestGetLatestUsedAtByUserIDs_UsesUsageLogs() {
	older := time.Now().Add(-4 * time.Hour).UTC().Truncate(time.Second)
	newer := time.Now().Add(-90 * time.Minute).UTC().Truncate(time.Second)

	userWithUsage := s.mustCreateUser(&service.User{Email: "usage-source@example.com"})
	userWithoutUsage := s.mustCreateUser(&service.User{Email: "usage-missing@example.com"})
	s.mustInsertUsageLog(userWithUsage.ID, older)
	s.mustInsertUsageLog(userWithUsage.ID, newer)

	got, err := s.repo.GetLatestUsedAtByUserIDs(s.ctx, []int64{userWithUsage.ID, userWithoutUsage.ID})
	s.Require().NoError(err)
	s.Require().Contains(got, userWithUsage.ID)
	s.Require().NotContains(got, userWithoutUsage.ID)
	s.Require().NotNil(got[userWithUsage.ID])
	s.Require().True(got[userWithUsage.ID].Equal(newer))
}

func (s *UserRepoSuite) TestListWithFilters_SortByLastUsedAtDesc_UsesUsageLogsNotLastActiveAt() {
	lastUsedOlder := time.Now().Add(-6 * time.Hour).UTC().Truncate(time.Second)
	lastUsedNewer := time.Now().Add(-2 * time.Hour).UTC().Truncate(time.Second)
	lastActiveVeryRecent := time.Now().Add(-10 * time.Minute).UTC().Truncate(time.Second)

	nilUsage := s.mustCreateUser(&service.User{Email: "nil-last-used@example.com"})
	wrongSource := s.mustCreateUser(&service.User{
		Email:        "active-not-usage@example.com",
		LastActiveAt: &lastActiveVeryRecent,
	})
	rightSource := s.mustCreateUser(&service.User{Email: "usage-wins@example.com"})

	s.mustInsertUsageLog(wrongSource.ID, lastUsedOlder)
	s.mustInsertUsageLog(rightSource.ID, lastUsedNewer)

	users, _, err := s.repo.ListWithFilters(s.ctx, pagination.PaginationParams{
		Page:      1,
		PageSize:  10,
		SortBy:    "last_used_at",
		SortOrder: "desc",
	}, service.UserListFilters{})
	s.Require().NoError(err)
	s.Require().Len(users, 3)
	s.Require().Equal(rightSource.ID, users[0].ID)
	s.Require().Equal(wrongSource.ID, users[1].ID)
	s.Require().Equal(nilUsage.ID, users[2].ID)
}

func (s *UserRepoSuite) mustInsertUsedBalanceReset(userID int64, usedAt time.Time) {
	s.T().Helper()
	s.mustInsertUsedBalanceResetWithNext(userID, usedAt, time.Time{})
}

func (s *UserRepoSuite) mustInsertUsedBalanceResetWithNext(userID int64, usedAt, nextResetAt time.Time) {
	s.T().Helper()
	create := s.client.RedeemCode.Create().
		SetCode(fmt.Sprintf("RESET-%d-%d", userID, usedAt.UnixNano())).
		SetType(service.RedeemTypeBalanceReset).
		SetValue(50).
		SetStatus(service.StatusUsed).
		SetUsedBy(userID).
		SetUsedAt(usedAt.UTC()).
		SetNotes("").
		SetValidityDays(0)
	if !nextResetAt.IsZero() {
		create.SetNextResetAt(nextResetAt.UTC())
	}
	_, err := create.Save(s.ctx)
	s.Require().NoError(err)
}

func (s *UserRepoSuite) TestListWithFilters_SortByNextBalanceResetAtAsc() {
	older := time.Now().Add(-10 * 24 * time.Hour).UTC().Truncate(time.Second)
	newer := time.Now().Add(-2 * 24 * time.Hour).UTC().Truncate(time.Second)

	none := s.mustCreateUser(&service.User{Email: "no-reset@example.com"})
	early := s.mustCreateUser(&service.User{Email: "early-reset@example.com"})
	late := s.mustCreateUser(&service.User{Email: "late-reset@example.com"})
	s.mustInsertUsedBalanceReset(early.ID, older)
	s.mustInsertUsedBalanceReset(late.ID, newer)

	users, _, err := s.repo.ListWithFilters(s.ctx, pagination.PaginationParams{
		Page:      1,
		PageSize:  10,
		SortBy:    "next_balance_reset_at",
		SortOrder: "asc",
	}, service.UserListFilters{})
	s.Require().NoError(err)
	s.Require().Len(users, 3)
	s.Require().Equal(early.ID, users[0].ID)
	s.Require().Equal(late.ID, users[1].ID)
	s.Require().Equal(none.ID, users[2].ID)
}

func (s *UserRepoSuite) TestListWithFilters_BalanceResetDueOverdueAndDueSoon() {
	overdueUsed := time.Now().Add(-8 * 24 * time.Hour).UTC().Truncate(time.Second)
	dueSoonUsed := time.Now().Add(-6 * 24 * time.Hour).UTC().Truncate(time.Second)
	laterUsed := time.Now().Add(-24 * time.Hour).UTC().Truncate(time.Second)

	overdueUser := s.mustCreateUser(&service.User{Email: "overdue-reset@example.com"})
	dueSoonUser := s.mustCreateUser(&service.User{Email: "duesoon-reset@example.com"})
	laterUser := s.mustCreateUser(&service.User{Email: "later-reset@example.com"})
	_ = s.mustCreateUser(&service.User{Email: "never-reset@example.com"})

	s.mustInsertUsedBalanceReset(overdueUser.ID, overdueUsed)
	s.mustInsertUsedBalanceReset(dueSoonUser.ID, dueSoonUsed)
	s.mustInsertUsedBalanceReset(laterUser.ID, laterUsed)

	overdue, _, err := s.repo.ListWithFilters(s.ctx, pagination.PaginationParams{
		Page: 1, PageSize: 10, SortBy: "email", SortOrder: "asc",
	}, service.UserListFilters{BalanceResetDue: service.BalanceResetDueOverdue})
	s.Require().NoError(err)
	s.Require().Len(overdue, 1)
	s.Require().Equal(overdueUser.ID, overdue[0].ID)

	dueSoon, _, err := s.repo.ListWithFilters(s.ctx, pagination.PaginationParams{
		Page: 1, PageSize: 10, SortBy: "email", SortOrder: "asc",
	}, service.UserListFilters{BalanceResetDue: service.BalanceResetDueSoon})
	s.Require().NoError(err)
	s.Require().Len(dueSoon, 1)
	s.Require().Equal(dueSoonUser.ID, dueSoon[0].ID)
}

func (s *UserRepoSuite) TestListWithFilters_CustomNextResetAtOverridesUsedAtPlusSeven() {
	usedAt := time.Now().Add(-24 * time.Hour).UTC().Truncate(time.Second)
	overdueNext := time.Now().Add(-time.Hour).UTC().Truncate(time.Second)
	laterNext := time.Now().Add(10 * 24 * time.Hour).UTC().Truncate(time.Second)

	overdueUser := s.mustCreateUser(&service.User{Email: "custom-overdue@example.com"})
	laterUser := s.mustCreateUser(&service.User{Email: "custom-later@example.com"})
	s.mustInsertUsedBalanceResetWithNext(overdueUser.ID, usedAt, overdueNext)
	s.mustInsertUsedBalanceResetWithNext(laterUser.ID, usedAt, laterNext)

	overdue, _, err := s.repo.ListWithFilters(s.ctx, pagination.PaginationParams{
		Page: 1, PageSize: 10, SortBy: "email", SortOrder: "asc",
	}, service.UserListFilters{BalanceResetDue: service.BalanceResetDueOverdue})
	s.Require().NoError(err)
	s.Require().Len(overdue, 1)
	s.Require().Equal(overdueUser.ID, overdue[0].ID)

	sorted, _, err := s.repo.ListWithFilters(s.ctx, pagination.PaginationParams{
		Page: 1, PageSize: 10, SortBy: "next_balance_reset_at", SortOrder: "asc",
	}, service.UserListFilters{})
	s.Require().NoError(err)
	s.Require().GreaterOrEqual(len(sorted), 2)
	s.Require().Equal(overdueUser.ID, sorted[0].ID)
	s.Require().Equal(laterUser.ID, sorted[1].ID)
}

func TestUserRepoSortSuiteSmoke(_ *testing.T) {}
