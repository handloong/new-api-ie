package controller

import (
	"errors"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
)

// LdapLoginReturnToken exchanges directory credentials for the caller's fixed API token.
// POST /api/user/ldap-login-return-token
//
// It always performs a real directory bind, never a local password check, then
// provisions the local account when auto-registration is enabled and returns the
// account's fixed-name token. No session is created: directory users are not
// expected to use the web console, so their local accounts carry no password.
func LdapLoginReturnToken(c *gin.Context) {
	if !common.LdapEnabled {
		common.ApiErrorT(c, "LDAP login is not enabled")
		return
	}

	var req LoginRequest
	if err := common.DecodeJson(c.Request.Body, &req); err != nil {
		common.ApiErrorT(c, "Invalid parameters")
		return
	}

	username := strings.TrimSpace(req.Username)
	password := req.Password
	if common.PasswordLoginEncryptionEnabled {
		if req.PasswordEncrypted == "" || req.EncryptionKeyID == "" {
			common.ApiErrorT(c, "Invalid parameters")
			return
		}
		decrypted, err := common.DecryptPassword(req.PasswordEncrypted, req.EncryptionKeyID)
		if err != nil {
			common.ApiErrorT(c, "Invalid username or password")
			return
		}
		password = decrypted
	}
	if username == "" || password == "" {
		common.ApiErrorT(c, "Invalid parameters")
		return
	}

	entry, err := service.LdapAuthenticate(username, password)
	if err != nil {
		common.SysLog(common.LogText("ldap authentication failed for user %s: %s", username, err.Error()))
		common.ApiErrorT(c, ldapAuthErrorMessage(err))
		return
	}

	result, err := model.ProvisionLdapUser(
		username,
		entry.DisplayName,
		entry.Email,
		common.LdapAutoRegister,
		common.LdapAutoRegisterPlanId,
	)
	if err != nil {
		common.SysLog(common.LogText("ldap user provisioning failed for %s: %s", username, err.Error()))
		common.ApiErrorT(c, ldapProvisionErrorMessage(err))
		return
	}

	token, err := model.GetOrCreateLdapToken(result.User.Id, result.User.Group)
	if err != nil {
		common.SysLog(common.LogText("ldap token issue failed for user %s: %s", username, err.Error()))
		common.ApiErrorT(c, "Failed to issue token")
		return
	}

	common.ApiSuccess(c, gin.H{
		"user_id":         result.User.Id,
		"username":        result.User.Username,
		"display_name":    result.User.DisplayName,
		"email":           result.User.Email,
		"group":           result.User.Group,
		"token_id":        token.Id,
		"name":            token.Name,
		"token":           token.Key,
		"expired_time":    token.ExpiredTime,
		"unlimited_quota": token.UnlimitedQuota,
		"remain_quota":    token.RemainQuota,
		"created":         result.Created,
	})
}

// ldapAuthErrorMessage maps internal directory errors to caller-facing text.
// Directory internals are deliberately collapsed into one generic message so the
// endpoint cannot be used to enumerate accounts or probe the directory.
func ldapAuthErrorMessage(err error) string {
	switch {
	case errors.Is(err, service.ErrLdapUserNotFound),
		errors.Is(err, service.ErrLdapInvalidCredentials):
		return "Invalid username or password"
	case errors.Is(err, service.ErrLdapAmbiguousUser):
		return "Multiple directory accounts match this username, please contact the administrator"
	case errors.Is(err, service.ErrLdapNotConfigured):
		return "LDAP service is not configured"
	default:
		return "LDAP authentication service is unavailable"
	}
}

func ldapProvisionErrorMessage(err error) string {
	switch {
	case errors.Is(err, model.ErrLdapAccountNotRegistered):
		return "This account is not registered, please contact the administrator"
	case errors.Is(err, model.ErrLdapAccountDisabled):
		return "This account has been disabled, please contact the administrator"
	case errors.Is(err, model.ErrLdapUsernameTaken):
		return "This username is held by a deleted account, please contact the administrator"
	default:
		return "Login failed, please try again later"
	}
}
