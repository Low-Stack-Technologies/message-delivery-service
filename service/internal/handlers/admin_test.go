package handlers

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Low-Stack-Technologies/message-delivery-service/internal/auth"
	"github.com/Low-Stack-Technologies/message-delivery-service/internal/config"
	"github.com/Low-Stack-Technologies/message-delivery-service/internal/state"
	"github.com/Low-Stack-Technologies/message-delivery-service/pkg/api"
	"github.com/go-chi/chi/v5"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"golang.org/x/crypto/ssh"
)

type fakeEmailSender struct {
	calls []emailCall
	err   error
}

type emailCall struct {
	From    string
	To      []string
	Subject string
	Body    string
	IsHTML  bool
}

func (f *fakeEmailSender) Send(from string, to []string, subject string, body string, isHTML bool) error {
	f.calls = append(f.calls, emailCall{
		From:    from,
		To:      append([]string(nil), to...),
		Subject: subject,
		Body:    body,
		IsHTML:  isHTML,
	})
	return f.err
}

type fakeSmsSender struct {
	calls []smsCall
	err   error
}

type smsCall struct {
	From string
	To   []string
	Body string
}

func (f *fakeSmsSender) Send(from string, to []string, body string) error {
	f.calls = append(f.calls, smsCall{
		From: from,
		To:   append([]string(nil), to...),
		Body: body,
	})
	return f.err
}

func TestAdminRoutes_WorkAgainstSQLiteStore(t *testing.T) {
	t.Cleanup(config.Reset)
	t.Setenv("ADMIN_BEARER_TOKEN", "")

	store := openAdminTestStore(t)
	t.Cleanup(func() {
		_ = store.Close()
	})

	_, creds, err := store.CreateAdminUser(context.Background(), "ui-admin")
	if err != nil {
		t.Fatalf("CreateAdminUser() failed: %v", err)
	}

	handler, _ := newAdminTestHandler(store)
	router := newAdminTestRouter(handler, store)

	loginResp := decodeAdminLoginResponse(t, doRequestJSON(t, router, http.MethodPost, "/v3/admin/auth/login", "", api.AdminLoginRequest{
		Username: "ui-admin",
		Password: creds.Password,
		TotpCode: currentTOTPCode(t, creds.TotpSecret),
	}))
	token := loginResp.Token

	meResp := decodeAdminMeResponse(t, doRequest(t, router, http.MethodGet, "/v3/admin/auth/me", token, nil))
	if meResp.User.Username != "ui-admin" {
		t.Fatalf("unexpected current admin user: %s", meResp.User.Username)
	}

	dashboard := decodeAdminDashboard(t, doRequest(t, router, http.MethodGet, "/v3/admin/dashboard", token, nil))
	if dashboard.Summary.Services != 3 {
		t.Fatalf("expected 3 services, got %d", dashboard.Summary.Services)
	}
	if dashboard.Summary.EmailAccounts != 2 {
		t.Fatalf("expected 2 email accounts, got %d", dashboard.Summary.EmailAccounts)
	}
	if dashboard.Summary.QueuedMessages != 1 {
		t.Fatalf("expected 1 queued message, got %d", dashboard.Summary.QueuedMessages)
	}

	restrictedMode := api.Restricted
	createServiceReq := api.AdminServiceCreateRequest{
		AllowedEmailAccountIds: &[]string{"support"},
		EmailAccessMode:        &restrictedMode,
		Id:                     "reports-api",
		Name:                   "Reports API",
		Owner:                  stringPtr("Analytics"),
		PublicKey:              "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIReportsApiExampleKey",
		Scope:                  api.AdminServiceScopeAll,
		Status:                 api.Active,
		Notes:                  stringPtr("Reporting consumer"),
	}
	createServiceResp := decodeAdminServiceResponse(t, doRequestJSON(t, router, http.MethodPost, "/v3/admin/services", token, createServiceReq))
	if createServiceResp.Service.Id != "reports-api" {
		t.Fatalf("unexpected service id: %q", createServiceResp.Service.Id)
	}
	if createServiceResp.Service.EmailAccessMode != api.Restricted {
		t.Fatalf("unexpected email access mode: %s", createServiceResp.Service.EmailAccessMode)
	}
	if len(createServiceResp.Service.AllowedEmailAccountIds) != 1 || createServiceResp.Service.AllowedEmailAccountIds[0] != "support" {
		t.Fatalf("unexpected allowed accounts: %#v", createServiceResp.Service.AllowedEmailAccountIds)
	}

	updateName := "Reports Platform"
	updateStatus := api.Paused
	updateServiceReq := api.AdminServiceUpdateRequest{
		Name:   &updateName,
		Status: &updateStatus,
	}
	updateServiceResp := decodeAdminServiceResponse(t, doRequestJSON(t, router, http.MethodPut, "/v3/admin/services/reports-api", token, updateServiceReq))
	if updateServiceResp.Service.Name != updateName {
		t.Fatalf("unexpected service name: %q", updateServiceResp.Service.Name)
	}
	if updateServiceResp.Service.Status != api.Paused {
		t.Fatalf("unexpected service status: %s", updateServiceResp.Service.Status)
	}
	if updateServiceResp.Service.EmailAccessMode != api.Restricted {
		t.Fatalf("expected email access mode to remain restricted, got %s", updateServiceResp.Service.EmailAccessMode)
	}
	if len(updateServiceResp.Service.AllowedEmailAccountIds) != 1 || updateServiceResp.Service.AllowedEmailAccountIds[0] != "support" {
		t.Fatalf("unexpected allowed accounts after update: %#v", updateServiceResp.Service.AllowedEmailAccountIds)
	}

	beforeReroll := updateServiceResp.Service.PublicKey
	rerollResp := decodeAdminServiceRerollResponse(t, doRequestJSON(t, router, http.MethodPost, "/v3/admin/services/reports-api/reroll", token, map[string]any{}))
	if rerollResp.PrivateKey == "" {
		t.Fatal("expected reroll private key")
	}
	if rerollResp.Service.PublicKey == beforeReroll {
		t.Fatal("expected reroll to change public key")
	}

	if rec := doRequest(t, router, http.MethodDelete, "/v3/admin/services/reports-api", token, nil); rec.Code != http.StatusNoContent {
		t.Fatalf("expected delete to succeed, got %d", rec.Code)
	}

	emailCreate := api.AdminEmailAccountCreateRequest{
		Address:     openapi_types.Email("ops@example.com"),
		DisplayName: stringPtr("Operations"),
		Id:          "ops",
		IsDefault:   boolPtr(true),
		Smtp: api.AdminEmailSmtpConfig{
			Host:     "smtp.example.com",
			Password: "secret",
			Port:     587,
			Username: "ops@example.com",
		},
	}
	emailResp := decodeAdminEmailAccountResponse(t, doRequestJSON(t, router, http.MethodPost, "/v3/admin/email-accounts", token, emailCreate))
	if emailResp.EmailAccount.Id != "ops" {
		t.Fatalf("unexpected email account id: %q", emailResp.EmailAccount.Id)
	}

	newAddress := openapi_types.Email("ops-updated@example.com")
	newDisplayName := "Support Desk"
	newDefault := false
	emailUpdate := api.AdminEmailAccountUpdateRequest{
		Address:     &newAddress,
		DisplayName: &newDisplayName,
		IsDefault:   &newDefault,
	}
	emailUpdateResp := decodeAdminEmailAccountResponse(t, doRequestJSON(t, router, http.MethodPut, "/v3/admin/email-accounts/ops", token, emailUpdate))
	if emailUpdateResp.EmailAccount.Address != newAddress {
		t.Fatalf("unexpected email address: %s", emailUpdateResp.EmailAccount.Address)
	}

	emailTestResp := decodeAdminEmailAccountTestResponse(t, doRequestJSON(t, router, http.MethodPost, "/v3/admin/email-accounts/ops/test", token, map[string]any{}))
	if emailTestResp.EmailAccount.LastTestedAt == nil {
		t.Fatal("expected email test timestamp")
	}

	if rec := doRequest(t, router, http.MethodDelete, "/v3/admin/email-accounts/ops", token, nil); rec.Code != http.StatusNoContent {
		t.Fatalf("expected email delete success, got %d", rec.Code)
	}

	smsResp := decodeAdminSmsCredentialsResponse(t, doRequestJSON(t, router, http.MethodPut, "/v3/admin/sms-credentials", token, api.AdminSmsCredentialsUpdateRequest{
		Username: "new-user",
		Password: "new-pass",
	}))
	if smsResp.SmsCredentials.Username != "new-user" {
		t.Fatalf("unexpected sms username: %q", smsResp.SmsCredentials.Username)
	}
	rotatedResp := decodeAdminSmsCredentialsRotateResponse(t, doRequestJSON(t, router, http.MethodPost, "/v3/admin/sms-credentials/rotate", token, map[string]any{}))
	if rotatedResp.SmsCredentials.RotationCount <= smsResp.SmsCredentials.RotationCount {
		t.Fatal("expected rotation count to increase")
	}

	usersResp := decodeAdminUsersResponse(t, doRequest(t, router, http.MethodGet, "/v3/admin/users", token, nil))
	if len(usersResp.Users) < 2 {
		t.Fatalf("expected at least 2 users, got %d", len(usersResp.Users))
	}

	createUserResp := decodeAdminUserCreateResponse(t, doRequestJSON(t, router, http.MethodPost, "/v3/admin/users", token, api.AdminUserCreateRequest{
		Username: "ops-admin",
	}))
	if createUserResp.User.Username != "ops-admin" {
		t.Fatalf("unexpected created admin user: %s", createUserResp.User.Username)
	}
	if createUserResp.Credentials.Password == "" || createUserResp.Credentials.TotpSecret == "" {
		t.Fatal("expected returned admin credentials")
	}

	previewReq := api.AdminMessageRequest{
		Channel:     api.AdminMessageChannelSms,
		ContentMode: api.Template,
		Recipients:  []string{"+46700000000"},
		SenderName:  stringPtr("AlertOps"),
		ServiceId:   "alerts-worker",
		Template: &api.AdminMessageTemplate{
			Name: "incident-sms",
			Data: map[string]interface{}{"incident": "INC-1001"},
		},
	}
	previewResp := decodeAdminMessagePreviewResponse(t, doRequestJSON(t, router, http.MethodPost, "/v3/admin/messages/preview", token, previewReq))
	if !strings.Contains(previewResp.Preview.Rendered, "incident-sms") {
		t.Fatalf("unexpected preview render: %s", previewResp.Preview.Rendered)
	}

	queueResp := decodeAdminMessageSubmitResponse(t, doRequestJSON(t, router, http.MethodPost, "/v3/admin/messages", token, previewReq))
	if queueResp.QueuedMessage.Id == "" {
		t.Fatal("expected queued message id")
	}

	dashboardAfter := decodeAdminDashboard(t, doRequest(t, router, http.MethodGet, "/v3/admin/dashboard", token, nil))
	if dashboardAfter.Summary.QueuedMessages != 2 {
		t.Fatalf("expected 2 queued messages after submit, got %d", dashboardAfter.Summary.QueuedMessages)
	}
}

func TestPublicSendEndpoints_UseSignedAuthAndFakeSenders(t *testing.T) {
	t.Cleanup(config.Reset)
	t.Setenv("ADMIN_BEARER_TOKEN", "")

	store := openAdminTestStore(t)
	t.Cleanup(func() {
		_ = store.Close()
	})

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey() failed: %v", err)
	}
	if err := store.UpdateConfig(func(cfg *config.Config) error {
		for i := range cfg.Services {
			if cfg.Services[i].ID != "billing-api" {
				continue
			}
			cfg.Services[i].PublicKey = string(ssh.MarshalAuthorizedKey(mustSSHPublicKey(t, pub)))
			cfg.Services[i].EmailAccessMode = "restricted"
			cfg.Services[i].AllowedEmailAccountIDs = []string{"support"}
		}
		return nil
	}); err != nil {
		t.Fatalf("UpdateConfig() failed: %v", err)
	}

	emailSender := &fakeEmailSender{}
	smsSender := &fakeSmsSender{}
	handler := NewHandler(emailSender, smsSender, store)
	router := newAdminTestRouter(handler, store)

	emailBody := []byte(`{"from":{"address":"support@example.com"},"subject":"Invoice ready","to":"finance@example.com","content":{"body":"Your invoice is ready for download.","isHtml":false}}`)
	emailRec := doSignedRequest(t, router, "/v3/email", emailBody, "billing-api", priv)
	if emailRec.Code != http.StatusAccepted {
		t.Fatalf("expected accepted email, got %d", emailRec.Code)
	}
	if len(emailSender.calls) != 1 {
		t.Fatalf("expected one email call, got %d", len(emailSender.calls))
	}
	if emailSender.calls[0].From != "support@example.com" {
		t.Fatalf("unexpected email sender: %s", emailSender.calls[0].From)
	}

	disallowedBody := []byte(`{"from":{"address":"receipts@example.com"},"subject":"Invoice ready","to":"finance@example.com","content":{"body":"Your invoice is ready for download.","isHtml":false}}`)
	disallowedRec := doSignedRequest(t, router, "/v3/email", disallowedBody, "billing-api", priv)
	if disallowedRec.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden email, got %d: %s", disallowedRec.Code, disallowedRec.Body.String())
	}
	if len(emailSender.calls) != 1 {
		t.Fatalf("expected forbidden sender to be blocked before send, got %d calls", len(emailSender.calls))
	}

	smsBody := []byte(`{"senderName":"AlertOps","to":"+46700000000","content":{"body":"Incident"}}`)
	smsRec := doSignedRequest(t, router, "/v3/sms", smsBody, "billing-api", priv)
	if smsRec.Code != http.StatusAccepted {
		t.Fatalf("expected accepted sms, got %d", smsRec.Code)
	}
	if len(smsSender.calls) != 1 {
		t.Fatalf("expected one sms call, got %d", len(smsSender.calls))
	}
	if smsSender.calls[0].From != "AlertOps" {
		t.Fatalf("unexpected sms sender: %s", smsSender.calls[0].From)
	}
}

func newAdminTestHandler(store *state.Store) (*Handler, *fakeEmailSender) {
	emailSender := &fakeEmailSender{}
	smsSender := &fakeSmsSender{}
	return NewHandler(emailSender, smsSender, store), emailSender
}

func newAdminTestRouter(handler *Handler, store *state.Store) chi.Router {
	r := chi.NewRouter()
	r.Use(auth.NewMiddleware(store))
	api.HandlerWithOptions(handler, api.ChiServerOptions{BaseRouter: r})
	return r
}

func openAdminTestStore(t *testing.T) *state.Store {
	t.Helper()

	dir := t.TempDir()
	store, err := state.New(filepath.Join(dir, "mds.db"))
	if err != nil {
		t.Fatalf("state.New() failed: %v", err)
	}
	return store
}

func doRequest(t *testing.T, router http.Handler, method, path, token string, body []byte) *httptest.ResponseRecorder {
	t.Helper()

	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		reader = bytes.NewReader(body)
	}

	req := httptest.NewRequest(method, path, reader)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func doRequestJSON(t *testing.T, router http.Handler, method, path, token string, value any) *httptest.ResponseRecorder {
	t.Helper()

	var body []byte
	if value != nil {
		body = mustJSON(t, value)
	}
	return doRequest(t, router, method, path, token, body)
}

func doSignedRequest(t *testing.T, router http.Handler, path string, body []byte, clientID string, priv ed25519.PrivateKey) *httptest.ResponseRecorder {
	t.Helper()

	timestamp := time.Now().UTC().Format(time.RFC3339)
	canonical := fmt.Sprintf("%s\n%s\n%s\n%x", http.MethodPost, path, timestamp, sha256Body(body))
	signature := ed25519.Sign(priv, []byte(canonical))

	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	req.Header.Set("X-Client-Id", clientID)
	req.Header.Set("X-Timestamp", timestamp)
	req.Header.Set("Authorization", "Signature "+base64Encode(signature))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func mustJSON(t *testing.T, value any) []byte {
	t.Helper()

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("json.Marshal() failed: %v", err)
	}
	return data
}

func decodeAdminDashboard(t *testing.T, rec *httptest.ResponseRecorder) api.AdminDashboardResponse {
	t.Helper()

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp api.AdminDashboardResponse
	decodeJSON(t, rec.Body.Bytes(), &resp)
	return resp
}

func decodeAdminServiceResponse(t *testing.T, rec *httptest.ResponseRecorder) api.AdminServiceResponse {
	t.Helper()

	if rec.Code != http.StatusOK && rec.Code != http.StatusCreated {
		t.Fatalf("expected success, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp api.AdminServiceResponse
	decodeJSON(t, rec.Body.Bytes(), &resp)
	return resp
}

func decodeAdminServiceRerollResponse(t *testing.T, rec *httptest.ResponseRecorder) api.AdminServiceRerollResponse {
	t.Helper()

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp api.AdminServiceRerollResponse
	decodeJSON(t, rec.Body.Bytes(), &resp)
	return resp
}

func decodeAdminEmailAccountResponse(t *testing.T, rec *httptest.ResponseRecorder) api.AdminEmailAccountResponse {
	t.Helper()

	if rec.Code != http.StatusOK && rec.Code != http.StatusCreated {
		t.Fatalf("expected success, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp api.AdminEmailAccountResponse
	decodeJSON(t, rec.Body.Bytes(), &resp)
	return resp
}

func decodeAdminEmailAccountTestResponse(t *testing.T, rec *httptest.ResponseRecorder) api.AdminEmailAccountTestResponse {
	t.Helper()

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp api.AdminEmailAccountTestResponse
	decodeJSON(t, rec.Body.Bytes(), &resp)
	return resp
}

func decodeAdminSmsCredentialsResponse(t *testing.T, rec *httptest.ResponseRecorder) api.AdminSmsCredentialsResponse {
	t.Helper()

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp api.AdminSmsCredentialsResponse
	decodeJSON(t, rec.Body.Bytes(), &resp)
	return resp
}

func decodeAdminSmsCredentialsRotateResponse(t *testing.T, rec *httptest.ResponseRecorder) api.AdminSmsCredentialsRotateResponse {
	t.Helper()

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp api.AdminSmsCredentialsRotateResponse
	decodeJSON(t, rec.Body.Bytes(), &resp)
	return resp
}

func decodeAdminLoginResponse(t *testing.T, rec *httptest.ResponseRecorder) api.AdminLoginResponse {
	t.Helper()

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp api.AdminLoginResponse
	decodeJSON(t, rec.Body.Bytes(), &resp)
	return resp
}

func decodeAdminMeResponse(t *testing.T, rec *httptest.ResponseRecorder) api.AdminMeResponse {
	t.Helper()

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp api.AdminMeResponse
	decodeJSON(t, rec.Body.Bytes(), &resp)
	return resp
}

func decodeAdminUsersResponse(t *testing.T, rec *httptest.ResponseRecorder) api.AdminUserListResponse {
	t.Helper()

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp api.AdminUserListResponse
	decodeJSON(t, rec.Body.Bytes(), &resp)
	return resp
}

func decodeAdminUserCreateResponse(t *testing.T, rec *httptest.ResponseRecorder) api.AdminUserCreateResponse {
	t.Helper()

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp api.AdminUserCreateResponse
	decodeJSON(t, rec.Body.Bytes(), &resp)
	return resp
}

func currentTOTPCode(t *testing.T, secret string) string {
	t.Helper()

	// The admin TOTP generator uses 30-second windows and SHA1.
	buf := time.Now().UTC().Unix() / 30
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(strings.TrimSpace(secret)))
	if err != nil {
		t.Fatalf("failed to decode TOTP secret: %v", err)
	}

	var counter [8]byte
	binary.BigEndian.PutUint64(counter[:], uint64(buf))
	mac := hmac.New(sha1.New, key)
	if _, err := mac.Write(counter[:]); err != nil {
		t.Fatalf("failed to build TOTP code: %v", err)
	}
	sum := mac.Sum(nil)
	offset := sum[len(sum)-1] & 0x0f
	code := (int(sum[offset])&0x7f)<<24 | (int(sum[offset+1])&0xff)<<16 | (int(sum[offset+2])&0xff)<<8 | (int(sum[offset+3]) & 0xff)
	return fmt.Sprintf("%06d", code%1000000)
}

func decodeAdminMessagePreviewResponse(t *testing.T, rec *httptest.ResponseRecorder) api.AdminMessagePreviewResponse {
	t.Helper()

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp api.AdminMessagePreviewResponse
	decodeJSON(t, rec.Body.Bytes(), &resp)
	return resp
}

func decodeAdminMessageSubmitResponse(t *testing.T, rec *httptest.ResponseRecorder) api.AdminMessageSubmitResponse {
	t.Helper()

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp api.AdminMessageSubmitResponse
	decodeJSON(t, rec.Body.Bytes(), &resp)
	return resp
}

func decodeJSON(t *testing.T, data []byte, target any) {
	t.Helper()

	if err := json.Unmarshal(data, target); err != nil {
		t.Fatalf("json.Unmarshal() failed: %v\nbody: %s", err, string(data))
	}
}

func sha256Body(body []byte) []byte {
	sum := sha256.Sum256(body)
	return sum[:]
}

func base64Encode(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func mustSSHPublicKey(t *testing.T, pub ed25519.PublicKey) ssh.PublicKey {
	t.Helper()

	key, err := ssh.NewPublicKey(pub)
	if err != nil {
		t.Fatalf("ssh.NewPublicKey() failed: %v", err)
	}
	return key
}
