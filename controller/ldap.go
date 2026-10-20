package controller

import (
	"errors"
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting/operation_setting"
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

// ldapRedeemRequest 是独立兑换页提交的三项内容：目录账号、目录密码、兑换码。
type ldapRedeemRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Key      string `json:"key"`
}

// LdapRedeem authenticates directory credentials and redeems a code for that account
// in a single request. It never creates a session and never issues an API token, so
// the redemption page cannot be used to obtain durable credentials.
// POST /api/user/ldap-redeem
//
// 该接口只服务独立兑换页，不经过控制台 i18n，因此 message 直接返回中文，并用 code
// 区分失败阶段，页面无需依赖翻译表。日志只记录用户名与失败原因，不记录密码和兑换码。
func LdapRedeem(c *gin.Context) {
	if !common.LdapEnabled {
		ldapRedeemFail(c, http.StatusForbidden, "ldap_disabled", "本站未启用 LDAP 登录")
		return
	}
	if !operation_setting.IsPaymentComplianceConfirmed() {
		ldapRedeemFail(c, http.StatusForbidden, "compliance_not_confirmed", "兑换功能尚未启用，请联系管理员")
		return
	}
	// 明文密码提交无法满足加密登录站点的要求，直接拒绝而不是绕过该保护。
	if common.PasswordLoginEncryptionEnabled {
		ldapRedeemFail(c, http.StatusForbidden, "encryption_required", "本站要求加密登录，请在控制台完成兑换")
		return
	}

	var req ldapRedeemRequest
	if err := common.DecodeJson(c.Request.Body, &req); err != nil {
		ldapRedeemFail(c, http.StatusBadRequest, "invalid_params", "参数格式错误")
		return
	}
	username := strings.TrimSpace(req.Username)
	key := strings.TrimSpace(req.Key)
	if username == "" || req.Password == "" || key == "" {
		ldapRedeemFail(c, http.StatusBadRequest, "missing_field", "账号、密码、兑换码均不能为空")
		return
	}

	entry, err := service.LdapAuthenticate(username, req.Password)
	if err != nil {
		common.SysLog(common.LogText("ldap redeem authentication failed for user %s: %s", username, err.Error()))
		switch {
		case errors.Is(err, service.ErrLdapAmbiguousUser):
			ldapRedeemFail(c, http.StatusUnauthorized, "ldap_ambiguous", "目录中存在多个同名账号，请联系管理员")
		case errors.Is(err, service.ErrLdapNotConfigured):
			ldapRedeemFail(c, http.StatusServiceUnavailable, "ldap_not_configured", "LDAP 服务未配置")
		case errors.Is(err, service.ErrLdapUserNotFound), errors.Is(err, service.ErrLdapInvalidCredentials):
			ldapRedeemFail(c, http.StatusUnauthorized, "ldap_auth_failed", "账号或密码错误")
		default:
			ldapRedeemFail(c, http.StatusServiceUnavailable, "ldap_unavailable", "LDAP 服务暂不可用，请稍后再试")
		}
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
		common.SysLog(common.LogText("ldap redeem provisioning failed for %s: %s", username, err.Error()))
		switch {
		case errors.Is(err, model.ErrLdapAccountNotRegistered):
			ldapRedeemFail(c, http.StatusForbidden, "account_not_registered", "账号未开通，请联系管理员")
		case errors.Is(err, model.ErrLdapAccountDisabled):
			ldapRedeemFail(c, http.StatusForbidden, "account_disabled", "账号已禁用，请联系管理员")
		case errors.Is(err, model.ErrLdapUsernameTaken):
			ldapRedeemFail(c, http.StatusForbidden, "username_taken", "用户名被已删除账号占用，请联系管理员")
		default:
			ldapRedeemFail(c, http.StatusInternalServerError, "provision_failed", "账号校验失败，请稍后再试")
		}
		return
	}

	// 兑换码的失败原因对调用方保持统一，避免据此判断兑换码是否存在、是否已使用。
	quota, err := model.Redeem(key, result.User.Id)
	if err != nil {
		common.SysLog(common.LogText("ldap redeem failed for user %s: %s", username, err.Error()))
		ldapRedeemFail(c, http.StatusBadRequest, "redeem_failed", "兑换失败，兑换码可能无效、已使用或已过期")
		return
	}

	// 顺带返回兑换后余额，页面不需要再发第二个请求。
	remainQuota, quotaErr := model.GetUserQuota(result.User.Id, false)
	if quotaErr != nil {
		common.SysLog(common.LogText("failed to read quota for user %s: %s", username, quotaErr.Error()))
		remainQuota = 0
	}

	common.ApiSuccess(c, gin.H{
		"username":       result.User.Username,
		"display_name":   result.User.DisplayName,
		"redeemed_quota": quota,
		"redeemed_usd":   quotaToDisplayAmount(int64(quota)),
		"remain_quota":   remainQuota,
		"remain_usd":     quotaToDisplayAmount(int64(remainQuota)),
		"created":        result.Created,
	})
}

func ldapRedeemFail(c *gin.Context, status int, code string, message string) {
	c.JSON(status, gin.H{
		"success": false,
		"code":    code,
		"message": message,
	})
}
