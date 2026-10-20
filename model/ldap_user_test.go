package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func withQuotaForNewUser(t *testing.T, quota int) {
	t.Helper()
	prev := common.QuotaForNewUser
	common.QuotaForNewUser = quota
	t.Cleanup(func() { common.QuotaForNewUser = prev })
}

func TestProvisionLdapUserCreatesPasswordlessAccount(t *testing.T) {
	truncateTables(t)
	withQuotaForNewUser(t, 1000)

	result, err := ProvisionLdapUser("zhangsan", "张三", "zs@corp.local", true, 0)
	require.NoError(t, err)
	require.True(t, result.Created)
	require.NotNil(t, result.User)

	assert.Empty(t, result.User.Password, "LDAP accounts must not carry a usable password")
	assert.Empty(t, result.Subscription, "no plan configured means no subscription")
	assert.Equal(t, 1000, result.User.Quota)
	assert.Equal(t, "default", result.User.Group)
	assert.Equal(t, "张三", result.User.DisplayName)
	assert.Equal(t, "zs@corp.local", result.User.Email)
}

func TestProvisionLdapUserBindsPlanAndZeroesNewUserQuota(t *testing.T) {
	truncateTables(t)
	withQuotaForNewUser(t, 1000)

	plan := &SubscriptionPlan{
		Title:         "LDAP Standard",
		DurationUnit:  SubscriptionDurationMonth,
		DurationValue: 1,
		TotalAmount:   500000,
		Enabled:       true,
	}
	plan.NormalizeDefaults()
	require.NoError(t, DB.Create(plan).Error)

	result, err := ProvisionLdapUser("lisi", "李四", "", true, plan.Id)
	require.NoError(t, err)
	require.True(t, result.Created)
	require.NotNil(t, result.Subscription)

	assert.Equal(t, plan.Id, result.Subscription.PlanId)
	assert.Equal(t, int64(500000), result.Subscription.AmountTotal)
	assert.Equal(t, 0, result.User.Quota, "the plan replaces the legacy new-user grant")

	var stored User
	require.NoError(t, DB.First(&stored, result.User.Id).Error)
	assert.Equal(t, 0, stored.Quota)
	assert.Empty(t, stored.Password)
}

func TestProvisionLdapUserReusesExistingAccount(t *testing.T) {
	truncateTables(t)
	withQuotaForNewUser(t, 0)

	first, err := ProvisionLdapUser("wangwu", "王五", "", true, 0)
	require.NoError(t, err)
	require.True(t, first.Created)

	second, err := ProvisionLdapUser("wangwu", "王五", "", true, 0)
	require.NoError(t, err)
	assert.False(t, second.Created)
	assert.Equal(t, first.User.Id, second.User.Id)

	var count int64
	require.NoError(t, DB.Model(&User{}).Where("username = ?", "wangwu").Count(&count).Error)
	assert.Equal(t, int64(1), count)
}

func TestProvisionLdapUserRequiresAutoRegister(t *testing.T) {
	truncateTables(t)

	_, err := ProvisionLdapUser("nobody", "", "", false, 0)
	assert.ErrorIs(t, err, ErrLdapAccountNotRegistered)
}

func TestProvisionLdapUserRejectsDisabledAccount(t *testing.T) {
	truncateTables(t)

	disabled := &User{
		Username: "disabled_user",
		Password: "",
		Status:   common.UserStatusDisabled,
		Group:    "default",
	}
	require.NoError(t, DB.Create(disabled).Error)

	_, err := ProvisionLdapUser("disabled_user", "", "", true, 0)
	assert.ErrorIs(t, err, ErrLdapAccountDisabled)
}

func TestLdapAccountCannotLoginWithPassword(t *testing.T) {
	truncateTables(t)
	withQuotaForNewUser(t, 0)

	_, err := ProvisionLdapUser("sunqi", "", "", true, 0)
	require.NoError(t, err)

	probe := User{Username: "sunqi", Password: "anything"}
	assert.Error(t, probe.ValidateAndFill(), "password login must be rejected for accounts without a password")
}

func TestGetOrCreateLdapTokenReusesToken(t *testing.T) {
	truncateTables(t)
	withQuotaForNewUser(t, 0)

	provisioned, err := ProvisionLdapUser("zhaoliu", "", "", true, 0)
	require.NoError(t, err)

	first, err := GetOrCreateLdapToken(provisioned.User.Id, "default")
	require.NoError(t, err)
	second, err := GetOrCreateLdapToken(provisioned.User.Id, "default")
	require.NoError(t, err)

	assert.Equal(t, first.Id, second.Id)
	assert.Equal(t, first.Key, second.Key)
	assert.Equal(t, LdapTokenName, first.Name)
	assert.True(t, first.UnlimitedQuota)
	assert.True(t, first.AllowUserQuotaQuery)
	assert.Equal(t, int64(-1), first.ExpiredTime, "the dsh token never expires")

	var count int64
	require.NoError(t, DB.Model(&Token{}).Where("user_id = ? AND name = ?", provisioned.User.Id, LdapTokenName).Count(&count).Error)
	assert.Equal(t, int64(1), count)
}
