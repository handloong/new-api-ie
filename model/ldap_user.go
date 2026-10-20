package model

import (
	"errors"
	"fmt"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

var (
	ErrLdapAccountDisabled      = errors.New("ldap: local account is disabled")
	ErrLdapAccountNotRegistered = errors.New("ldap: account is not registered")
	ErrLdapUsernameTaken        = errors.New("ldap: username is held by a deleted account")
)

// LdapProvisionResult reports what provisioning did for a directory user.
type LdapProvisionResult struct {
	User         *User
	Created      bool
	Subscription *UserSubscription
}

// GetUserByUsername returns a non soft-deleted user by exact username.
func GetUserByUsername(username string) (*User, error) {
	var user User
	if err := DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// ProvisionLdapUser resolves a directory user to a local account.
//
// Existing accounts are returned untouched. Unknown accounts are created only
// when autoRegister is true. Created accounts store an empty password, which is
// this project's "no password" representation: prepareForInsert skips hashing,
// ValidateAndFill rejects the account, and the console login stays closed to
// directory users by design.
//
// When planId > 0 the new account is bound to that subscription plan inside the
// same transaction, and the legacy QuotaForNewUser grant is zeroed because the
// plan's TotalAmount is meant to be the only quota source for such users.
func ProvisionLdapUser(username, displayName, email string, autoRegister bool, planId int) (*LdapProvisionResult, error) {
	user, err := GetUserByUsername(username)
	if err == nil {
		if user.Status != common.UserStatusEnabled {
			return nil, ErrLdapAccountDisabled
		}
		return &LdapProvisionResult{User: user}, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if !autoRegister {
		return nil, ErrLdapAccountNotRegistered
	}

	// A soft-deleted account still holds the username unique index, so creating a
	// fresh row for the same name would fail. Surface it instead of guessing.
	taken, err := CheckUserExistOrDeleted(username, "")
	if err != nil {
		return nil, err
	}
	if taken {
		return nil, ErrLdapUsernameTaken
	}

	result := &LdapProvisionResult{Created: true}
	err = DB.Transaction(func(tx *gorm.DB) error {
		newUser := &User{
			Username:    username,
			Password:    "",
			DisplayName: truncateDisplayName(displayName, username),
			Email:       email,
			Role:        common.RoleCommonUser,
			Status:      common.UserStatusEnabled,
			Group:       "default",
		}
		if err := newUser.InsertWithTx(tx, 0); err != nil {
			return err
		}
		result.User = newUser

		if planId <= 0 {
			return nil
		}
		plan, err := getSubscriptionPlanByIdTx(tx, planId)
		if err != nil {
			return fmt.Errorf("ldap: load subscription plan %d: %w", planId, err)
		}
		subscription, err := CreateUserSubscriptionFromPlanTx(tx, newUser.Id, plan, "ldap")
		if err != nil {
			return fmt.Errorf("ldap: bind subscription plan %d: %w", planId, err)
		}
		// InsertWithTx granted QuotaForNewUser; the plan replaces that grant.
		if err := tx.Model(&User{}).Where("id = ?", newUser.Id).Update("quota", 0).Error; err != nil {
			return err
		}
		newUser.Quota = 0
		result.Subscription = subscription
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// truncateDisplayName keeps the display name within the 20-rune column limit and
// falls back to the username when the directory returns nothing useful.
func truncateDisplayName(displayName string, username string) string {
	name := displayName
	if name == "" {
		name = username
	}
	runes := []rune(name)
	if len(runes) > 20 {
		runes = runes[:20]
	}
	return string(runes)
}
