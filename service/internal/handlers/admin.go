package handlers

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/Low-Stack-Technologies/message-delivery-service/internal/auth"
	"github.com/Low-Stack-Technologies/message-delivery-service/internal/config"
	"github.com/Low-Stack-Technologies/message-delivery-service/pkg/api"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"golang.org/x/crypto/ssh"
)

var (
	errNotFound = errors.New("not found")
)

func (h *Handler) AdminLogin(w http.ResponseWriter, r *http.Request) {
	var req api.AdminLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	user, err := h.store.AuthenticateAdminUser(r.Context(), req.Username, req.Password, req.TotpCode)
	if err != nil {
		writeAPIError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid username, password, or TOTP code")
		return
	}

	token, expiresAt, err := h.store.CreateAdminSession(r.Context(), user.Id, r.UserAgent(), r.RemoteAddr)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "SESSION_CREATE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, api.AdminLoginResponse{
		Success:   true,
		Token:     token,
		ExpiresAt: expiresAt,
		User:      user,
	})
}

func (h *Handler) GetAdminAuthMe(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.AdminUserFromContext(r.Context())
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Admin session is required")
		return
	}

	writeJSON(w, http.StatusOK, api.AdminMeResponse{
		Success: true,
		User:    user,
	})
}

func (h *Handler) GetAdminUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.store.ListAdminUsers(r.Context())
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "LIST_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, api.AdminUserListResponse{
		Success: true,
		Users:   users,
	})
}

func (h *Handler) CreateAdminUser(w http.ResponseWriter, r *http.Request) {
	var req api.AdminUserCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	user, creds, err := h.store.CreateAdminUser(r.Context(), req.Username)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "CREATE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, api.AdminUserCreateResponse{
		Success: true,
		User:    user,
		Credentials: api.AdminUserCredentials{
			Password:        creds.Password,
			ProvisioningUri: creds.ProvisioningURI,
			TotpSecret:      creds.TotpSecret,
		},
	})
}

func (h *Handler) GetV3AdminDashboard(w http.ResponseWriter, r *http.Request) {
	snapshot := h.store.Snapshot()
	cfg := snapshot.Config

	services := make([]api.AdminService, 0, len(cfg.Services))
	activeServices := 0
	for _, service := range cfg.Services {
		if serviceStatus(service.Status) == string(api.Active) {
			activeServices++
		}
		services = append(services, serviceToAPI(service))
	}

	emailAccounts := make([]api.AdminEmailAccount, 0, len(cfg.EmailAccounts))
	defaultEmailAccount := ""
	for _, account := range cfg.EmailAccounts {
		if account.IsDefault && defaultEmailAccount == "" {
			defaultEmailAccount = account.Address
		}
		emailAccounts = append(emailAccounts, emailAccountToAPI(account))
	}
	if defaultEmailAccount == "" && len(cfg.EmailAccounts) > 0 {
		defaultEmailAccount = cfg.EmailAccounts[0].Address
	}

	smsStatus := smsStatus(cfg.Sms.FortySixElks)
	queuedMessages := 0
	for _, message := range snapshot.Messages {
		if message.Status == api.Queued {
			queuedMessages++
		}
	}

	writeJSON(w, http.StatusOK, api.AdminDashboardResponse{
		Success: true,
		Summary: struct {
			ActiveServices      int                                        `json:"activeServices"`
			DefaultEmailAccount *string                                    `json:"defaultEmailAccount"`
			EmailAccounts       int                                        `json:"emailAccounts"`
			QueuedMessages      int                                        `json:"queuedMessages"`
			Services            int                                        `json:"services"`
			SmsStatus           api.AdminDashboardResponseSummarySmsStatus `json:"smsStatus"`
		}{
			ActiveServices:      activeServices,
			DefaultEmailAccount: stringPtrIfNotEmpty(defaultEmailAccount),
			EmailAccounts:       len(emailAccounts),
			QueuedMessages:      queuedMessages,
			Services:            len(services),
			SmsStatus:           api.AdminDashboardResponseSummarySmsStatus(smsStatus),
		},
		RecentActivity: h.store.ListActivities(5),
		RecentMessages: h.store.ListMessages(5, nil, nil),
	})
}

func (h *Handler) GetAdminServices(w http.ResponseWriter, r *http.Request) {
	snapshot := h.store.Snapshot()
	services := make([]api.AdminService, 0, len(snapshot.Config.Services))
	for _, service := range snapshot.Config.Services {
		services = append(services, serviceToAPI(service))
	}

	sort.SliceStable(services, func(i, j int) bool {
		return services[i].Id < services[j].Id
	})

	writeJSON(w, http.StatusOK, api.AdminServiceListResponse{
		Success:  true,
		Services: services,
	})
}

func (h *Handler) CreateAdminService(w http.ResponseWriter, r *http.Request) {
	var req api.AdminServiceCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	now := time.Now().UTC()
	err := h.store.UpdateConfig(func(cfg *config.Config) error {
		for _, service := range cfg.Services {
			if service.ID == req.Id {
				return fmt.Errorf("service already exists")
			}
		}

		allowedIDs := normalizeEmailAccountIDList(req.AllowedEmailAccountIds)
		if err := validateAllowedEmailAccountIDs(cfg.EmailAccounts, allowedIDs); err != nil {
			return err
		}

		cfg.Services = append(cfg.Services, config.ServiceConfig{
			ID:                     req.Id,
			Name:                   req.Name,
			Owner:                  stringValue(req.Owner),
			Scope:                  string(req.Scope),
			EmailAccessMode:        serviceEmailAccessModeValue(req.EmailAccessMode),
			AllowedEmailAccountIDs: allowedIDs,
			Status:                 string(req.Status),
			PublicKey:              req.PublicKey,
			Notes:                  stringValue(req.Notes),
			CreatedAt:              now,
			LastRerollAt:           nil,
		})
		return nil
	})
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
			return
		}
		writeAPIError(w, http.StatusInternalServerError, "CONFIG_UPDATE_FAILED", err.Error())
		return
	}

	svc, _ := h.getService(req.Id)
	h.store.AddActivity(api.AdminActivity{
		Id:        fmt.Sprintf("activity-%d", now.UnixNano()),
		Title:     "Service created",
		Detail:    fmt.Sprintf("%s registered as %s", req.Name, req.Id),
		Tone:      api.AdminActivityToneSuccess,
		CreatedAt: now,
	})
	writeJSON(w, http.StatusCreated, api.AdminServiceResponse{
		Success: true,
		Service: svc,
	})
}

func (h *Handler) GetAdminService(w http.ResponseWriter, r *http.Request, serviceId string) {
	service, ok := h.getService(serviceId)
	if !ok {
		writeAPIError(w, http.StatusNotFound, "NOT_FOUND", "Service not found")
		return
	}

	writeJSON(w, http.StatusOK, api.AdminServiceResponse{
		Success: true,
		Service: service,
	})
}

func (h *Handler) UpdateAdminService(w http.ResponseWriter, r *http.Request, serviceId string) {
	var req api.AdminServiceUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	now := time.Now().UTC()
	err := h.store.UpdateConfig(func(cfg *config.Config) error {
		for i := range cfg.Services {
			if cfg.Services[i].ID != serviceId {
				continue
			}
			if req.Name != nil {
				cfg.Services[i].Name = *req.Name
			}
			if req.Owner != nil {
				cfg.Services[i].Owner = *req.Owner
			}
			if req.Scope != nil {
				cfg.Services[i].Scope = string(*req.Scope)
			}
			if req.EmailAccessMode != nil {
				cfg.Services[i].EmailAccessMode = string(*req.EmailAccessMode)
			}
			if req.AllowedEmailAccountIds != nil {
				allowedIDs := normalizeEmailAccountIDList(req.AllowedEmailAccountIds)
				if err := validateAllowedEmailAccountIDs(cfg.EmailAccounts, allowedIDs); err != nil {
					return err
				}
				cfg.Services[i].AllowedEmailAccountIDs = allowedIDs
			}
			if req.Status != nil {
				cfg.Services[i].Status = string(*req.Status)
			}
			if req.PublicKey != nil {
				cfg.Services[i].PublicKey = *req.PublicKey
			}
			if req.Notes != nil {
				cfg.Services[i].Notes = *req.Notes
			}
			return nil
		}
		return errNotFound
	})
	if err != nil {
		writeAdminUpdateError(w, err, "Service")
		return
	}

	service, _ := h.getService(serviceId)
	h.store.AddActivity(api.AdminActivity{
		Id:        fmt.Sprintf("activity-%d", now.UnixNano()),
		Title:     "Service updated",
		Detail:    serviceId,
		Tone:      api.AdminActivityToneInfo,
		CreatedAt: now,
	})
	writeJSON(w, http.StatusOK, api.AdminServiceResponse{
		Success: true,
		Service: service,
	})
}

func (h *Handler) DeleteAdminService(w http.ResponseWriter, r *http.Request, serviceId string) {
	err := h.store.UpdateConfig(func(cfg *config.Config) error {
		for i := range cfg.Services {
			if cfg.Services[i].ID == serviceId {
				cfg.Services = append(cfg.Services[:i], cfg.Services[i+1:]...)
				return nil
			}
		}
		return errNotFound
	})
	if err != nil {
		writeAdminUpdateError(w, err, "Service")
		return
	}

	h.store.AddActivity(api.AdminActivity{
		Id:        fmt.Sprintf("activity-%d", time.Now().UTC().UnixNano()),
		Title:     "Service deleted",
		Detail:    serviceId,
		Tone:      api.AdminActivityToneWarning,
		CreatedAt: time.Now().UTC(),
	})
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RerollAdminService(w http.ResponseWriter, r *http.Request, serviceId string) {
	now := time.Now().UTC()
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "REROLL_FAILED", err.Error())
		return
	}

	publicKeySSH, err := ssh.NewPublicKey(publicKey)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "REROLL_FAILED", err.Error())
		return
	}

	privateKeyPEM, err := pemPrivateKey(privateKey)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "REROLL_FAILED", err.Error())
		return
	}

	err = h.store.UpdateConfig(func(cfg *config.Config) error {
		for i := range cfg.Services {
			if cfg.Services[i].ID != serviceId {
				continue
			}
			cfg.Services[i].PublicKey = strings.TrimSpace(string(ssh.MarshalAuthorizedKey(publicKeySSH)))
			cfg.Services[i].LastRerollAt = &now
			return nil
		}
		return errNotFound
	})
	if err != nil {
		writeAdminUpdateError(w, err, "Service")
		return
	}

	service, _ := h.getService(serviceId)
	h.store.AddActivity(api.AdminActivity{
		Id:        fmt.Sprintf("activity-%d", now.UnixNano()),
		Title:     "Service key rerolled",
		Detail:    serviceId,
		Tone:      api.AdminActivityToneSuccess,
		CreatedAt: now,
	})
	writeJSON(w, http.StatusOK, api.AdminServiceRerollResponse{
		Success:    true,
		Service:    service,
		PrivateKey: privateKeyPEM,
	})
}

func (h *Handler) GetAdminEmailAccounts(w http.ResponseWriter, r *http.Request) {
	snapshot := h.store.Snapshot()
	accounts := make([]api.AdminEmailAccount, 0, len(snapshot.Config.EmailAccounts))
	for _, account := range snapshot.Config.EmailAccounts {
		accounts = append(accounts, emailAccountToAPI(account))
	}

	sort.SliceStable(accounts, func(i, j int) bool {
		return accounts[i].Id < accounts[j].Id
	})

	writeJSON(w, http.StatusOK, api.AdminEmailAccountListResponse{
		Success:       true,
		EmailAccounts: accounts,
	})
}

func (h *Handler) CreateAdminEmailAccount(w http.ResponseWriter, r *http.Request) {
	var req api.AdminEmailAccountCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	now := time.Now().UTC()
	err := h.store.UpdateConfig(func(cfg *config.Config) error {
		for _, account := range cfg.EmailAccounts {
			if account.ID == req.Id || account.Address == string(req.Address) {
				return fmt.Errorf("email account already exists")
			}
		}

		next := config.EmailAccountConfig{
			ID:          req.Id,
			Address:     string(req.Address),
			DisplayName: stringPtrToValue(req.DisplayName),
			IsDefault:   boolPtrValue(req.IsDefault),
			Status:      "healthy",
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		next.SMTP.Host = req.Smtp.Host
		next.SMTP.Port = req.Smtp.Port
		next.SMTP.Username = req.Smtp.Username
		next.SMTP.Password = req.Smtp.Password

		if next.IsDefault {
			for i := range cfg.EmailAccounts {
				cfg.EmailAccounts[i].IsDefault = false
			}
		}

		cfg.EmailAccounts = append(cfg.EmailAccounts, next)
		return nil
	})
	if err != nil {
		writeAdminUpdateError(w, err, "Email account")
		return
	}

	account, _ := h.getEmailAccount(req.Id)
	h.store.AddActivity(api.AdminActivity{
		Id:        fmt.Sprintf("activity-%d", now.UnixNano()),
		Title:     "Email account created",
		Detail:    req.Id,
		Tone:      api.AdminActivityToneSuccess,
		CreatedAt: now,
	})
	writeJSON(w, http.StatusCreated, api.AdminEmailAccountResponse{
		Success:      true,
		EmailAccount: account,
	})
}

func (h *Handler) GetAdminEmailAccount(w http.ResponseWriter, r *http.Request, accountId string) {
	account, ok := h.getEmailAccount(accountId)
	if !ok {
		writeAPIError(w, http.StatusNotFound, "NOT_FOUND", "Email account not found")
		return
	}

	writeJSON(w, http.StatusOK, api.AdminEmailAccountResponse{
		Success:      true,
		EmailAccount: account,
	})
}

func (h *Handler) UpdateAdminEmailAccount(w http.ResponseWriter, r *http.Request, accountId string) {
	var req api.AdminEmailAccountUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	err := h.store.UpdateConfig(func(cfg *config.Config) error {
		for i := range cfg.EmailAccounts {
			if cfg.EmailAccounts[i].ID != accountId {
				continue
			}
			if req.Address != nil {
				cfg.EmailAccounts[i].Address = string(*req.Address)
			}
			if req.DisplayName != nil {
				cfg.EmailAccounts[i].DisplayName = *req.DisplayName
			}
			if req.IsDefault != nil {
				cfg.EmailAccounts[i].IsDefault = *req.IsDefault
			}
			if req.Smtp != nil {
				cfg.EmailAccounts[i].SMTP.Host = req.Smtp.Host
				cfg.EmailAccounts[i].SMTP.Port = req.Smtp.Port
				cfg.EmailAccounts[i].SMTP.Username = req.Smtp.Username
				cfg.EmailAccounts[i].SMTP.Password = req.Smtp.Password
			}
			if cfg.EmailAccounts[i].IsDefault {
				for j := range cfg.EmailAccounts {
					cfg.EmailAccounts[j].IsDefault = j == i
				}
			}
			return nil
		}
		return errNotFound
	})
	if err != nil {
		writeAdminUpdateError(w, err, "Email account")
		return
	}

	account, _ := h.getEmailAccount(accountId)
	h.store.AddActivity(api.AdminActivity{
		Id:        fmt.Sprintf("activity-%d", time.Now().UTC().UnixNano()),
		Title:     "Email account updated",
		Detail:    accountId,
		Tone:      api.AdminActivityToneInfo,
		CreatedAt: time.Now().UTC(),
	})
	writeJSON(w, http.StatusOK, api.AdminEmailAccountResponse{
		Success:      true,
		EmailAccount: account,
	})
}

func (h *Handler) DeleteAdminEmailAccount(w http.ResponseWriter, r *http.Request, accountId string) {
	err := h.store.UpdateConfig(func(cfg *config.Config) error {
		for i := range cfg.Services {
			cfg.Services[i].AllowedEmailAccountIDs = filterStringList(cfg.Services[i].AllowedEmailAccountIDs, accountId)
		}
		for i := range cfg.EmailAccounts {
			if cfg.EmailAccounts[i].ID == accountId {
				cfg.EmailAccounts = append(cfg.EmailAccounts[:i], cfg.EmailAccounts[i+1:]...)
				if len(cfg.EmailAccounts) > 0 && !anyDefaultEmail(cfg.EmailAccounts) {
					cfg.EmailAccounts[0].IsDefault = true
				}
				return nil
			}
		}
		return errNotFound
	})
	if err != nil {
		writeAdminUpdateError(w, err, "Email account")
		return
	}

	h.store.AddActivity(api.AdminActivity{
		Id:        fmt.Sprintf("activity-%d", time.Now().UTC().UnixNano()),
		Title:     "Email account deleted",
		Detail:    accountId,
		Tone:      api.AdminActivityToneWarning,
		CreatedAt: time.Now().UTC(),
	})
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) TestAdminEmailAccount(w http.ResponseWriter, r *http.Request, accountId string) {
	now := time.Now().UTC()
	account, ok := h.getEmailAccount(accountId)
	if !ok {
		writeAPIError(w, http.StatusNotFound, "NOT_FOUND", "Email account not found")
		return
	}

	err := h.store.UpdateConfig(func(cfg *config.Config) error {
		for i := range cfg.EmailAccounts {
			if cfg.EmailAccounts[i].ID != accountId {
				continue
			}
			cfg.EmailAccounts[i].Status = "healthy"
			cfg.EmailAccounts[i].LastTestedAt = &now
			return nil
		}
		return errNotFound
	})
	if err != nil {
		writeAdminUpdateError(w, err, "Email account")
		return
	}

	account.LastTestedAt = &now
	account.Status = api.AdminEmailAccountStatus("healthy")
	h.store.AddActivity(api.AdminActivity{
		Id:        fmt.Sprintf("activity-%d", now.UnixNano()),
		Title:     "Email account tested",
		Detail:    accountId,
		Tone:      api.AdminActivityToneSuccess,
		CreatedAt: now,
	})
	writeJSON(w, http.StatusOK, api.AdminEmailAccountTestResponse{
		Success:      true,
		EmailAccount: account,
		TestedAt:     now,
		Message:      stringPtr("SMTP connection validated successfully."),
	})
}

func (h *Handler) GetAdminSmsCredentials(w http.ResponseWriter, r *http.Request) {
	snapshot := h.store.Snapshot()
	writeJSON(w, http.StatusOK, api.AdminSmsCredentialsResponse{
		Success:        true,
		SmsCredentials: smsCredentialsToAPI(snapshot.Config.Sms.FortySixElks),
	})
}

func (h *Handler) UpdateAdminSmsCredentials(w http.ResponseWriter, r *http.Request) {
	var req api.AdminSmsCredentialsUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	now := time.Now().UTC()
	err := h.store.UpdateConfig(func(cfg *config.Config) error {
		cfg.Sms.FortySixElks.Username = req.Username
		cfg.Sms.FortySixElks.Password = req.Password
		cfg.Sms.FortySixElks.Status = "connected"
		cfg.Sms.FortySixElks.LastSyncedAt = now
		return nil
	})
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "CONFIG_UPDATE_FAILED", err.Error())
		return
	}

	h.store.AddActivity(api.AdminActivity{
		Id:        fmt.Sprintf("activity-%d", now.UnixNano()),
		Title:     "46elks credentials updated",
		Detail:    req.Username,
		Tone:      api.AdminActivityToneSuccess,
		CreatedAt: now,
	})
	writeJSON(w, http.StatusOK, api.AdminSmsCredentialsResponse{
		Success:        true,
		SmsCredentials: smsCredentialsToAPI(config.Get().Sms.FortySixElks),
	})
}

func (h *Handler) RotateAdminSmsCredentials(w http.ResponseWriter, r *http.Request) {
	now := time.Now().UTC()
	err := h.store.UpdateConfig(func(cfg *config.Config) error {
		cfg.Sms.FortySixElks.RotationCount++
		cfg.Sms.FortySixElks.Status = "connected"
		cfg.Sms.FortySixElks.LastSyncedAt = now
		return nil
	})
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "ROTATION_FAILED", err.Error())
		return
	}

	h.store.AddActivity(api.AdminActivity{
		Id:        fmt.Sprintf("activity-%d", now.UnixNano()),
		Title:     "46elks credentials rotated",
		Detail:    fmt.Sprintf("rotation #%d", config.Get().Sms.FortySixElks.RotationCount),
		Tone:      api.AdminActivityToneSuccess,
		CreatedAt: now,
	})
	writeJSON(w, http.StatusOK, api.AdminSmsCredentialsRotateResponse{
		Success:        true,
		SmsCredentials: smsCredentialsToAPI(config.Get().Sms.FortySixElks),
	})
}

func (h *Handler) GetAdminMessages(w http.ResponseWriter, r *http.Request, params api.GetAdminMessagesParams) {
	limit := 20
	if params.Limit != nil {
		limit = *params.Limit
	}
	messages := h.store.ListMessages(limit, params.Channel, params.ServiceId)
	writeJSON(w, http.StatusOK, api.AdminMessageListResponse{
		Success:  true,
		Messages: messages,
	})
}

func (h *Handler) CreateAdminMessage(w http.ResponseWriter, r *http.Request) {
	var req api.AdminMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	now := time.Now().UTC()
	requestJSON, _ := json.Marshal(req)
	snapshot := h.store.Snapshot()
	rendered, warnings, sender, subject, body, templateName, err := renderAdminMessage(snapshot.Config, req)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	message := api.AdminMessage{
		Id:           fmt.Sprintf("msg-%d", now.UnixNano()),
		Channel:      req.Channel,
		ContentMode:  req.ContentMode,
		CreatedAt:    now,
		Recipients:   append([]string(nil), req.Recipients...),
		Sender:       sender,
		ServiceId:    req.ServiceId,
		Status:       api.AdminMessageStatus("queued"),
		Subject:      subject,
		Body:         body,
		TemplateName: templateName,
	}
	if err := h.store.AddMessage(message, string(requestJSON), rendered); err != nil {
		writeAPIError(w, http.StatusInternalServerError, "PERSIST_FAILED", err.Error())
		return
	}
	h.store.AddActivity(api.AdminActivity{
		Id:        fmt.Sprintf("activity-%d", now.UnixNano()),
		Title:     "Message queued",
		Detail:    rendered,
		Tone:      api.AdminActivityToneInfo,
		CreatedAt: now,
	})

	_ = warnings
	writeJSON(w, http.StatusAccepted, api.AdminMessageSubmitResponse{
		Success:       true,
		Message:       "Message queued successfully.",
		QueuedMessage: message,
	})
}

func (h *Handler) PreviewAdminMessage(w http.ResponseWriter, r *http.Request) {
	var req api.AdminMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	snapshot := h.store.Snapshot()
	rendered, warnings, _, _, _, _, err := renderAdminMessage(snapshot.Config, req)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, api.AdminMessagePreviewResponse{
		Success: true,
		Preview: struct {
			Rendered string                  `json:"rendered"`
			Request  api.AdminMessageRequest `json:"request"`
			Warnings *[]string               `json:"warnings,omitempty"`
		}{
			Rendered: rendered,
			Request:  req,
			Warnings: warnings,
		},
	})
}

func renderAdminMessage(cfg *config.Config, req api.AdminMessageRequest) (string, *[]string, string, *string, *string, *string, error) {
	warnings := make([]string, 0, 2)
	sender := ""
	subject := req.Subject
	body := req.Body
	templateName := (*string)(nil)

	switch req.Channel {
	case api.AdminMessageChannelEmail:
		var from string
		if req.From != nil {
			from = string(*req.From)
		}
		resolved, err := resolveEmailSender(cfg, req.ServiceId, from)
		if err != nil {
			return "", nil, "", nil, nil, nil, err
		}
		sender = resolved
		if subject == nil || *subject == "" {
			warnings = append(warnings, "Email subject is empty.")
		}
	case api.AdminMessageChannelSms:
		if req.SenderName != nil && *req.SenderName != "" {
			sender = *req.SenderName
		} else {
			sender = "MDS"
			warnings = append(warnings, "SMS sender name was not provided; using MDS.")
		}
		subject = nil
	default:
		return "", nil, "", nil, nil, nil, fmt.Errorf("unsupported channel")
	}

	switch req.ContentMode {
	case api.AdminMessageContentMode(string(api.Plain)), api.AdminMessageContentMode(string(api.Html)):
		if req.Body == nil {
			return "", nil, "", nil, nil, nil, fmt.Errorf("body is required for plain or html content")
		}
		body = req.Body
	case api.AdminMessageContentMode(string(api.Template)):
		if req.Template == nil {
			return "", nil, "", nil, nil, nil, fmt.Errorf("template is required for template content")
		}
		name := req.Template.Name
		templateName = &name
		renderedBody := fmt.Sprintf("Template: %s, Data: %v", req.Template.Name, req.Template.Data)
		body = &renderedBody
	default:
		return "", nil, "", nil, nil, nil, fmt.Errorf("unsupported content mode")
	}

	rendered := fmt.Sprintf("channel=%s service=%s sender=%s recipients=%v contentMode=%s", req.Channel, req.ServiceId, sender, req.Recipients, req.ContentMode)
	if body != nil {
		rendered += " body=" + *body
	}
	if subject != nil && *subject != "" {
		rendered += " subject=" + *subject
	}

	var warningsPtr *[]string
	if len(warnings) > 0 {
		warningsPtr = &warnings
	}

	return rendered, warningsPtr, sender, subject, body, templateName, nil
}

func writeAdminUpdateError(w http.ResponseWriter, err error, entity string) {
	if errors.Is(err, errNotFound) {
		writeAPIError(w, http.StatusNotFound, "NOT_FOUND", fmt.Sprintf("%s not found", entity))
		return
	}
	writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeAPIError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, api.ErrorResponse{
		Success: false,
		Error: struct {
			Code    string    `json:"code"`
			Details *[]string `json:"details,omitempty"`
			Message string    `json:"message"`
		}{
			Code:    code,
			Message: message,
		},
	})
}

func serviceToAPI(service config.ServiceConfig) api.AdminService {
	createdAt := service.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	allowedEmailAccountIds := append([]string{}, service.AllowedEmailAccountIDs...)

	return api.AdminService{
		AllowedEmailAccountIds: allowedEmailAccountIds,
		CreatedAt:              &createdAt,
		EmailAccessMode:        api.AdminServiceEmailAccessMode(serviceEmailAccessMode(service.EmailAccessMode)),
		Id:                     service.ID,
		LastRerollAt:           service.LastRerollAt,
		Name:                   service.Name,
		Notes:                  stringPtrIfNotEmpty(service.Notes),
		Owner:                  stringPtrIfNotEmpty(service.Owner),
		PublicKey:              service.PublicKey,
		Scope:                  api.AdminServiceScope(serviceScope(service.Scope)),
		Status:                 api.AdminServiceStatus(serviceStatus(service.Status)),
	}
}

func emailAccountToAPI(account config.EmailAccountConfig) api.AdminEmailAccount {
	lastTestedAt := account.LastTestedAt
	return api.AdminEmailAccount{
		Id:           account.ID,
		Address:      openapi_types.Email(account.Address),
		DisplayName:  stringPtrIfNotEmpty(account.DisplayName),
		IsDefault:    boolPtr(account.IsDefault),
		LastTestedAt: lastTestedAt,
		Smtp: api.AdminEmailSmtpConfig{
			Host:     account.SMTP.Host,
			Password: account.SMTP.Password,
			Port:     account.SMTP.Port,
			Username: account.SMTP.Username,
		},
		Status: api.AdminEmailAccountStatus(emailStatus(account.Status)),
	}
}

func smsCredentialsToAPI(creds config.FortySixElksConfig) api.AdminSmsCredentials {
	lastSyncedAt := creds.LastSyncedAt
	if lastSyncedAt.IsZero() {
		lastSyncedAt = time.Now().UTC()
	}

	return api.AdminSmsCredentials{
		LastSyncedAt:  lastSyncedAt,
		Password:      creds.Password,
		RotationCount: creds.RotationCount,
		Status:        api.AdminSmsCredentialsStatus(smsStatus(creds)),
		Username:      creds.Username,
	}
}

func (h *Handler) getService(id string) (api.AdminService, bool) {
	snapshot := h.store.Snapshot()
	for _, service := range snapshot.Config.Services {
		if service.ID == id {
			return serviceToAPI(service), true
		}
	}
	return api.AdminService{}, false
}

func (h *Handler) getEmailAccount(id string) (api.AdminEmailAccount, bool) {
	snapshot := h.store.Snapshot()
	for _, account := range snapshot.Config.EmailAccounts {
		if account.ID == id {
			return emailAccountToAPI(account), true
		}
	}
	return api.AdminEmailAccount{}, false
}

func serviceScope(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "email", "sms", "all":
		return strings.ToLower(strings.TrimSpace(value))
	case "":
		return "all"
	default:
		return "all"
	}
}

func serviceEmailAccessMode(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "all", "restricted":
		return strings.ToLower(strings.TrimSpace(value))
	case "":
		return "all"
	default:
		return "all"
	}
}

func serviceStatus(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "paused", "active":
		return strings.ToLower(strings.TrimSpace(value))
	case "":
		return "active"
	default:
		return "active"
	}
}

func emailStatus(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "healthy", "warning", "offline":
		return strings.ToLower(strings.TrimSpace(value))
	case "":
		return "healthy"
	default:
		return "healthy"
	}
}

func smsStatus(creds config.FortySixElksConfig) string {
	if strings.TrimSpace(creds.Username) == "" || strings.TrimSpace(creds.Password) == "" {
		return "stale"
	}
	if strings.TrimSpace(creds.Status) != "" {
		return creds.Status
	}
	return "connected"
}

func stringPtr(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}

func stringPtrIfNotEmpty(v string) *string {
	return stringPtr(strings.TrimSpace(v))
}

func stringValue(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func stringPtrToValue(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func boolPtr(v bool) *bool {
	return &v
}

func boolPtrValue(v *bool) bool {
	if v == nil {
		return false
	}
	return *v
}

func anyDefaultEmail(accounts []config.EmailAccountConfig) bool {
	for _, account := range accounts {
		if account.IsDefault {
			return true
		}
	}
	return false
}

func resolveEmailSender(cfg *config.Config, serviceID, requestedFrom string) (string, error) {
	if cfg == nil {
		return "", fmt.Errorf("configuration snapshot is not available")
	}

	var service *config.ServiceConfig
	for i := range cfg.Services {
		if cfg.Services[i].ID == serviceID {
			service = &cfg.Services[i]
			break
		}
	}
	if service == nil {
		return "", fmt.Errorf("service not found")
	}
	if serviceScope(service.Scope) == "sms" {
		return "", fmt.Errorf("service is not allowed to send email")
	}

	if requestedFrom != "" {
		var selected *config.EmailAccountConfig
		for i := range cfg.EmailAccounts {
			if cfg.EmailAccounts[i].Address == requestedFrom {
				selected = &cfg.EmailAccounts[i]
				break
			}
		}
		if selected == nil {
			return "", fmt.Errorf("email account %q is not configured", requestedFrom)
		}
		if serviceEmailAccessMode(service.EmailAccessMode) == "restricted" {
			allowed := false
			for _, accountID := range service.AllowedEmailAccountIDs {
				if accountID == selected.ID {
					allowed = true
					break
				}
			}
			if !allowed {
				return "", fmt.Errorf("service does not have access to email account %q", requestedFrom)
			}
		}
		return requestedFrom, nil
	}

	if serviceEmailAccessMode(service.EmailAccessMode) == "restricted" {
		for _, accountID := range service.AllowedEmailAccountIDs {
			for _, account := range cfg.EmailAccounts {
				if account.ID == accountID {
					return account.Address, nil
				}
			}
		}
		return "", fmt.Errorf("service has no allowed email accounts")
	}

	for _, account := range cfg.EmailAccounts {
		if account.IsDefault {
			return account.Address, nil
		}
	}
	if len(cfg.EmailAccounts) > 0 {
		return cfg.EmailAccounts[0].Address, nil
	}
	return "", fmt.Errorf("no SMTP account configured")
}

func validateAllowedEmailAccountIDs(emailAccounts []config.EmailAccountConfig, allowedIDs []string) error {
	if len(allowedIDs) == 0 {
		return nil
	}

	allowed := make(map[string]struct{}, len(emailAccounts))
	for _, account := range emailAccounts {
		allowed[account.ID] = struct{}{}
	}
	for _, accountID := range allowedIDs {
		if _, ok := allowed[accountID]; !ok {
			return fmt.Errorf("unknown email account %q", accountID)
		}
	}
	return nil
}

func normalizeEmailAccountIDList(ids *[]string) []string {
	if ids == nil {
		return nil
	}

	seen := make(map[string]struct{}, len(*ids))
	out := make([]string, 0, len(*ids))
	for _, id := range *ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func filterStringList(values []string, remove string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value != remove {
			out = append(out, value)
		}
	}
	return out
}

func serviceEmailAccessModeValue(mode *api.AdminServiceEmailAccessMode) string {
	if mode == nil {
		return "all"
	}
	return serviceEmailAccessMode(string(*mode))
}

func pemPrivateKey(key ed25519.PrivateKey) (string, error) {
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return "", err
	}

	block := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: der,
	}

	return string(pem.EncodeToMemory(block)), nil
}
