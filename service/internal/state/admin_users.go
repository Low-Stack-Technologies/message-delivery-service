package state

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"log"
	"math/big"
	"net/url"
	"strings"
	"time"

	sqlcdb "github.com/Low-Stack-Technologies/message-delivery-service/internal/db/sqlc"
	"github.com/Low-Stack-Technologies/message-delivery-service/pkg/api"
	"golang.org/x/crypto/bcrypt"
)

const (
	AdminIssuer     = "Message Delivery Service"
	adminSessionTTL = 24 * time.Hour
	totpPeriod      = 30
	totpDigits      = 6
)

type AdminUserCredentials struct {
	Password       string
	TotpSecret     string
	ProvisioningURI string
}

func (s *Store) ensureDefaultAdminUser(ctx context.Context) error {
	count, err := s.q.CountAdminUsers(ctx)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	user, creds, err := s.createAdminUserTx(ctx, s.q, "admin")
	if err != nil {
		return err
	}

	log.Printf("Created default admin user credentials: username=%s password=%s totp_secret=%s totp_uri=%s", user.Username, creds.Password, creds.TotpSecret, creds.ProvisioningURI)
	return nil
}

func (s *Store) ListAdminUsers(ctx context.Context) ([]api.AdminUser, error) {
	rows, err := s.q.ListAdminUsers(ctx)
	if err != nil {
		return nil, err
	}
	return adminUsersToAPI(rows), nil
}

func (s *Store) CreateAdminUser(ctx context.Context, username string) (api.AdminUser, AdminUserCredentials, error) {
	return s.createAdminUserTx(ctx, s.q, username)
}

func (s *Store) AuthenticateAdminUser(ctx context.Context, username, password, totpCode string) (api.AdminUser, error) {
	row, err := s.q.GetAdminUserByUsername(ctx, username)
	if err != nil {
		return api.AdminUser{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(row.PasswordHash), []byte(password)); err != nil {
		return api.AdminUser{}, fmt.Errorf("invalid credentials")
	}

	if !verifyTOTP(row.TotpSecret, totpCode, time.Now().UTC()) {
		return api.AdminUser{}, fmt.Errorf("invalid credentials")
	}

	now := time.Now().UTC()
	if err := s.q.UpdateAdminUserLastLogin(ctx, sqlcdb.UpdateAdminUserLastLoginParams{
		LastLoginAt: &now,
		ID:          row.ID,
	}); err != nil {
		return api.AdminUser{}, err
	}

	return adminUserPublicToAPI(row.ID, row.Username, row.CreatedAt, row.UpdatedAt, &now), nil
}

func (s *Store) CreateAdminSession(ctx context.Context, userID string, userAgent, remoteAddr string) (string, time.Time, error) {
	token, hash, err := generateSessionToken()
	if err != nil {
		return "", time.Time{}, err
	}

	now := time.Now().UTC()
	expiresAt := now.Add(adminSessionTTL)
	if _, err := s.q.CreateAdminSession(ctx, sqlcdb.CreateAdminSessionParams{
		TokenHash:  hash,
		UserID:     userID,
		CreatedAt:  now,
		ExpiresAt:  expiresAt,
		LastUsedAt: now,
		RevokedAt:  nil,
		UserAgent:  nullableString(userAgent),
		RemoteAddr: nullableString(remoteAddr),
	}); err != nil {
		return "", time.Time{}, err
	}

	return token, expiresAt, nil
}

func (s *Store) ValidateAdminSession(ctx context.Context, token string) (api.AdminUser, error) {
	hash := hashToken(token)
	row, err := s.q.GetAdminSessionByTokenHash(ctx, hash)
	if err != nil {
		return api.AdminUser{}, err
	}

	now := time.Now().UTC()
	if row.ExpiresAt.Before(now) {
		return api.AdminUser{}, fmt.Errorf("session expired")
	}
	if row.RevokedAt != nil {
		return api.AdminUser{}, fmt.Errorf("session revoked")
	}

	if err := s.q.TouchAdminSession(ctx, sqlcdb.TouchAdminSessionParams{
		LastUsedAt: now,
		TokenHash:  hash,
	}); err != nil {
		return api.AdminUser{}, err
	}

	user, err := s.q.GetAdminUserPublicByID(ctx, row.UserID)
	if err != nil {
		return api.AdminUser{}, err
	}
	return adminUserPublicToAPI(user.ID, user.Username, user.CreatedAt, user.UpdatedAt, user.LastLoginAt), nil
}

func (s *Store) createAdminUserTx(ctx context.Context, q *sqlcdb.Queries, username string) (api.AdminUser, AdminUserCredentials, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return api.AdminUser{}, AdminUserCredentials{}, fmt.Errorf("username is required")
	}

	password, err := generatePassword()
	if err != nil {
		return api.AdminUser{}, AdminUserCredentials{}, err
	}
	secret, err := generateTOTPSecret()
	if err != nil {
		return api.AdminUser{}, AdminUserCredentials{}, err
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return api.AdminUser{}, AdminUserCredentials{}, err
	}

	now := time.Now().UTC()
	id := fmt.Sprintf("admin-%d", now.UnixNano())
	user, err := q.CreateAdminUser(ctx, sqlcdb.CreateAdminUserParams{
		ID:           id,
		Username:     username,
		PasswordHash: string(passwordHash),
		TotpSecret:   secret,
		CreatedAt:    now,
		UpdatedAt:    now,
		LastLoginAt:  nil,
	})
	if err != nil {
		return api.AdminUser{}, AdminUserCredentials{}, err
	}

	creds := AdminUserCredentials{
		Password:       password,
		TotpSecret:     secret,
		ProvisioningURI: provisioningURI(username, secret),
	}
	return adminUserPublicToAPI(user.ID, user.Username, user.CreatedAt, user.UpdatedAt, user.LastLoginAt), creds, nil
}

func adminUsersToAPI(rows []sqlcdb.ListAdminUsersRow) []api.AdminUser {
	items := make([]api.AdminUser, 0, len(rows))
	for _, row := range rows {
		items = append(items, adminUserPublicToAPI(row.ID, row.Username, row.CreatedAt, row.UpdatedAt, row.LastLoginAt))
	}
	return items
}

func adminUserPublicToAPI(id, username string, createdAt, updatedAt time.Time, lastLoginAt *time.Time) api.AdminUser {
	return api.AdminUser{
		Id:          id,
		Username:    username,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		LastLoginAt: lastLoginAt,
	}
}

func generatePassword() (string, error) {
	return randomString(20)
}

func generateTOTPSecret() (string, error) {
	buf := make([]byte, 20)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	enc := base32.StdEncoding.WithPadding(base32.NoPadding)
	return strings.ToUpper(enc.EncodeToString(buf)), nil
}

func provisioningURI(username, secret string) string {
	label := url.PathEscape(AdminIssuer + ":" + username)
	issuer := url.QueryEscape(AdminIssuer)
	return fmt.Sprintf("otpauth://totp/%s?secret=%s&issuer=%s&algorithm=SHA1&digits=%d&period=%d", label, secret, issuer, totpDigits, totpPeriod)
}

func verifyTOTP(secret, code string, now time.Time) bool {
	code = strings.TrimSpace(code)
	if code == "" {
		return false
	}
	for offset := int64(-1); offset <= 1; offset++ {
		candidate, err := totpCode(secret, now.Add(time.Duration(offset*totpPeriod)*time.Second))
		if err == nil && subtleEquals(candidate, code) {
			return true
		}
	}
	return false
}

func totpCode(secret string, now time.Time) (string, error) {
	counter := now.Unix() / totpPeriod
	mac, err := totpHMAC(secret, counter)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%0*d", totpDigits, mac%pow10(totpDigits)), nil
}

func totpHMAC(secret string, counter int64) (int, error) {
	dec, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(strings.TrimSpace(secret)))
	if err != nil {
		return 0, err
	}

	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], uint64(counter))

	mac := hmac.New(sha1.New, dec)
	if _, err := mac.Write(buf[:]); err != nil {
		return 0, err
	}
	sum := mac.Sum(nil)
	offset := sum[len(sum)-1] & 0x0f
	code := (int(sum[offset])&0x7f)<<24 | (int(sum[offset+1])&0xff)<<16 | (int(sum[offset+2])&0xff)<<8 | (int(sum[offset+3]) & 0xff)
	return code, nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func generateSessionToken() (string, string, error) {
	raw, err := randomString(48)
	if err != nil {
		return "", "", err
	}
	return raw, hashToken(raw), nil
}

func randomString(length int) (string, error) {
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var out strings.Builder
	out.Grow(length)
	max := big.NewInt(int64(len(alphabet)))
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		out.WriteByte(alphabet[n.Int64()])
	}
	return out.String(), nil
}

func pow10(n int) int {
	out := 1
	for i := 0; i < n; i++ {
		out *= 10
	}
	return out
}

func subtleEquals(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var v byte
	for i := range a {
		v |= a[i] ^ b[i]
	}
	return v == 0
}

func nullableString(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	v := value
	return &v
}
