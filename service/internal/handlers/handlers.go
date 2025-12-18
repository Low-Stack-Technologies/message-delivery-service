package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Low-Stack-Technologies/message-delivery-service/internal/config"
	"github.com/Low-Stack-Technologies/message-delivery-service/internal/delivery"
	"github.com/Low-Stack-Technologies/message-delivery-service/pkg/api"
)

type Handler struct {
	email *delivery.EmailProvider
	sms   *delivery.SmsProvider
}

func NewHandler(email *delivery.EmailProvider, sms *delivery.SmsProvider) *Handler {
	return &Handler{
		email: email,
		sms:   sms,
	}
}

func (h *Handler) GetHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func (h *Handler) PostV3Email(w http.ResponseWriter, r *http.Request, params api.PostV3EmailParams) {
	var req api.EmailRequest
	config.DebugLog("[DEBUG] PostV3Email - Decoding request body...")
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		config.DebugLog("[DEBUG] PostV3Email - Decode error: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 1. Extract Recipients
	var addresses []string
	if addr, err := req.To.AsEmailRequestTo0(); err == nil {
		addresses = append(addresses, string(addr))
	} else if contact, err := req.To.AsEmailContact(); err == nil {
		addresses = append(addresses, string(contact.Address))
	} else if multi, err := req.To.AsEmailRequestTo2(); err == nil {
		for _, item := range multi {
			if a, err := item.AsEmailRequestTo20(); err == nil {
				addresses = append(addresses, string(a))
			} else if c, err := item.AsEmailContact(); err == nil {
				addresses = append(addresses, string(c.Address))
			}
		}
	}

	if len(addresses) == 0 {
		config.DebugLog("[DEBUG] PostV3Email - No recipients extracted from: %+v", req.To)
		http.Error(w, "No recipients specified", http.StatusBadRequest)
		return
	}
	config.DebugLog("[DEBUG] PostV3Email - Recipients: %v", addresses)

	// 2. Extract Content
	var body string
	var isHTML bool
	if req.Content != nil {
		if c0, err := req.Content.AsEmailRequestContent0(); err == nil {
			body = c0.Body
			if c0.IsHtml != nil {
				isHTML = *c0.IsHtml
			}
		} else if c1, err := req.Content.AsEmailRequestContent1(); err == nil {
			// In a real app, we'd render the template here
			body = fmt.Sprintf("Template: %s, Data: %v", c1.Template.Name, c1.Template.Data)
		}
	}

	// 3. Send
	if err := h.email.Send(string(req.From.Address), addresses, req.Subject, body, isHTML); err != nil {
		log.Printf("Email send error: %v", err)
		h.sendError(w, "DELIVERY_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(api.SuccessResponse{
		Success: true,
		Message: "Email accepted for delivery",
	})
}

func (h *Handler) sendError(w http.ResponseWriter, code, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(api.ErrorResponse{
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

func (h *Handler) PostV3Sms(w http.ResponseWriter, r *http.Request, params api.PostV3SmsParams) {
	var req api.SmsRequest
	config.DebugLog("[DEBUG] PostV3Sms - Decoding request body...")
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		config.DebugLog("[DEBUG] PostV3Sms - Decode error: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 1. Extract Recipients
	var numbers []string
	if single, err := req.To.AsSmsRecipient(); err == nil {
		if s0, err := single.AsSmsRecipient0(); err == nil {
			numbers = append(numbers, s0)
		} else if s1, err := single.AsSmsRecipient1(); err == nil {
			numbers = append(numbers, s1.Phone) // Simplification: assuming phone is already E.164 or handled by provider
		}
	} else if multi, err := req.To.AsSmsRequestTo1(); err == nil {
		for _, item := range multi {
			if s0, err := item.AsSmsRecipient0(); err == nil {
				numbers = append(numbers, s0)
			} else if s1, err := item.AsSmsRecipient1(); err == nil {
				numbers = append(numbers, s1.Phone)
			}
		}
	}

	if len(numbers) == 0 {
		config.DebugLog("[DEBUG] PostV3Sms - No recipients extracted from: %+v", req.To)
		http.Error(w, "No recipients specified", http.StatusBadRequest)
		return
	}
	config.DebugLog("[DEBUG] PostV3Sms - Recipients: %v", numbers)

	// 2. Extract Content
	var body string
	if req.Content != nil {
		if c0, err := req.Content.AsSmsRequestContent0(); err == nil {
			body = c0.Body
		} else if c1, err := req.Content.AsSmsRequestContent1(); err == nil {
			body = fmt.Sprintf("Template: %s, Data: %v", c1.Template.Name, c1.Template.Data)
		}
	}

	// 3. Send
	if err := h.sms.Send(req.SenderName, numbers, body); err != nil {
		log.Printf("SMS send error: %v", err)
		h.sendError(w, "DELIVERY_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(api.SuccessResponse{
		Success: true,
		Message: "SMS accepted for delivery",
	})
}
