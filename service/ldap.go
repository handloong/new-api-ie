package service

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/go-ldap/ldap/v3"
)

var (
	ErrLdapUserNotFound       = errors.New("ldap: user not found")
	ErrLdapInvalidCredentials = errors.New("ldap: invalid credentials")
	ErrLdapAmbiguousUser      = errors.New("ldap: ambiguous user match")
	ErrLdapNotConfigured      = errors.New("ldap: server is not configured")
)

const (
	ldapDialTimeout  = 10 * time.Second
	ldapSearchLimit  = 10
	ldapAttrMail     = "mail"
	ldapDefaultPort  = 389
	ldapsDefaultPort = 636

	ldapDefaultUserFilter      = "(&(objectClass=user)(sAMAccountName={0}))"
	ldapDefaultDisplayNameAttr = "displayName"
)

// LdapUserEntry is the profile fetched from the directory after a successful bind.
type LdapUserEntry struct {
	DN          string
	DisplayName string
	Email       string
}

type ldapConfig struct {
	server          string
	port            int
	useTls          bool
	bindDn          string
	bindPassword    string
	baseDn          string
	userFilter      string
	displayNameAttr string
}

func currentLdapConfig() ldapConfig {
	displayNameAttr := strings.TrimSpace(common.LdapAttrDisplayName)
	if displayNameAttr == "" {
		displayNameAttr = ldapDefaultDisplayNameAttr
	}
	userFilter := strings.TrimSpace(common.LdapUserFilter)
	if userFilter == "" {
		userFilter = ldapDefaultUserFilter
	}
	return ldapConfig{
		server:          common.LdapServer,
		port:            common.LdapPort,
		useTls:          common.LdapUseTls,
		bindDn:          strings.TrimSpace(common.LdapBindDn),
		bindPassword:    common.LdapBindPassword,
		baseDn:          strings.TrimSpace(common.LdapBaseDn),
		userFilter:      userFilter,
		displayNameAttr: displayNameAttr,
	}
}

// LdapAuthenticate verifies a username/password pair against the directory.
//
// It binds with the service account, searches for the user by the configured
// filter, then re-binds with the matched DN and the supplied password, so the
// password is verified by the directory itself and never stored locally.
func LdapAuthenticate(username, password string) (*LdapUserEntry, error) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return nil, ErrLdapInvalidCredentials
	}

	cfg := currentLdapConfig()
	conn, err := dialLdap(cfg)
	if err != nil {
		return nil, err
	}
	defer func() { _ = conn.Close() }()

	// A blank service account means an anonymous bind, which only works when the
	// directory allows anonymous search.
	if cfg.bindDn != "" {
		if err = conn.Bind(cfg.bindDn, cfg.bindPassword); err != nil {
			return nil, fmt.Errorf("ldap: service account bind failed: %w", err)
		}
	}

	// {0} is replaced with the RFC4515-escaped username to prevent filter injection.
	filter := strings.ReplaceAll(cfg.userFilter, "{0}", ldap.EscapeFilter(username))
	searchRequest := ldap.NewSearchRequest(
		cfg.baseDn,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		ldapSearchLimit,
		int(ldapDialTimeout.Seconds()),
		false,
		filter,
		ldapSearchAttrs(cfg.displayNameAttr),
		nil,
	)
	result, err := conn.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("ldap: search user failed: %w", err)
	}
	switch len(result.Entries) {
	case 0:
		return nil, ErrLdapUserNotFound
	case 1:
		// The only acceptable case: exactly one match.
	default:
		return nil, ErrLdapAmbiguousUser
	}

	entry := result.Entries[0]
	if err = conn.Bind(entry.DN, password); err != nil {
		if ldap.IsErrorWithCode(err, ldap.LDAPResultInvalidCredentials) {
			return nil, ErrLdapInvalidCredentials
		}
		return nil, fmt.Errorf("ldap: user bind failed: %w", err)
	}

	return &LdapUserEntry{
		DN:          entry.DN,
		DisplayName: strings.TrimSpace(entry.GetAttributeValue(cfg.displayNameAttr)),
		Email:       strings.TrimSpace(entry.GetAttributeValue(ldapAttrMail)),
	}, nil
}

// ldapSearchAttrs returns the attributes to fetch, de-duplicated so that a
// display-name attribute configured as "mail" does not make the server reject
// the request for repeating an attribute.
func ldapSearchAttrs(displayNameAttr string) []string {
	attrs := make([]string, 0, 2)
	seen := make(map[string]struct{}, 2)
	for _, attr := range []string{displayNameAttr, ldapAttrMail} {
		if attr == "" {
			continue
		}
		if _, ok := seen[attr]; ok {
			continue
		}
		seen[attr] = struct{}{}
		attrs = append(attrs, attr)
	}
	return attrs
}

func dialLdap(cfg ldapConfig) (*ldap.Conn, error) {
	server := strings.TrimSpace(cfg.server)
	if server == "" {
		return nil, ErrLdapNotConfigured
	}
	if cfg.baseDn == "" {
		return nil, errors.New("ldap: base dn is not configured")
	}

	scheme := "ldap"
	if cfg.useTls {
		scheme = "ldaps"
	}
	lower := strings.ToLower(server)
	if strings.HasPrefix(lower, "ldaps://") {
		scheme = "ldaps"
		server = server[len("ldaps://"):]
	} else if strings.HasPrefix(lower, "ldap://") {
		server = server[len("ldap://"):]
	}
	server = strings.TrimSuffix(strings.TrimSpace(server), "/")

	if _, _, err := net.SplitHostPort(server); err != nil {
		port := cfg.port
		if port <= 0 {
			port = ldapDefaultPort
			if scheme == "ldaps" {
				port = ldapsDefaultPort
			}
		}
		server = net.JoinHostPort(server, strconv.Itoa(port))
	}

	conn, err := ldap.DialURL(
		fmt.Sprintf("%s://%s", scheme, server),
		ldap.DialWithDialer(&net.Dialer{Timeout: ldapDialTimeout}),
	)
	if err != nil {
		return nil, fmt.Errorf("ldap: connect failed: %w", err)
	}
	return conn, nil
}
