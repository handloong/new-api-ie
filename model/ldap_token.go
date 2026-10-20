package model

import (
	"errors"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

// LdapTokenName is the fixed token name handed out by the LDAP credential endpoint.
const LdapTokenName = "dsh"

// GetOrCreateLdapToken returns the user's fixed-name token, creating it on first
// use. The token never expires and has unlimited quota, so a directory client
// keeps the same credential across calls instead of accumulating new ones.
func GetOrCreateLdapToken(userId int, group string) (*Token, error) {
	token, err := findLdapToken(userId)
	if err == nil {
		return token, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	key, err := common.GenerateKey()
	if err != nil {
		return nil, err
	}
	candidate := &Token{
		UserId:              userId,
		Key:                 key,
		Status:              common.TokenStatusEnabled,
		Name:                LdapTokenName,
		CreatedTime:         common.GetTimestamp(),
		AccessedTime:        common.GetTimestamp(),
		ExpiredTime:         -1,
		UnlimitedQuota:      true,
		Group:               group,
		AllowUserQuotaQuery: true,
	}
	if err = candidate.Insert(); err != nil {
		// A concurrent request may have created the token first; reuse it rather
		// than surfacing a unique-constraint error.
		if existing, findErr := findLdapToken(userId); findErr == nil {
			return existing, nil
		}
		return nil, err
	}
	return candidate, nil
}

func findLdapToken(userId int) (*Token, error) {
	var token Token
	if err := DB.Where("user_id = ? AND name = ?", userId, LdapTokenName).First(&token).Error; err != nil {
		return nil, err
	}
	return &token, nil
}
