-- +migrate Up
INSERT INTO app_settings (id, server_host, server_port, debug, admin_bearer_token, updated_at)
VALUES (1, '0.0.0.0', 3000, 0, 'change-me', '2026-05-05 00:00:00');

INSERT INTO services (id, name, owner, scope, status, public_key, notes, created_at, last_reroll_at, updated_at) VALUES
('billing-api', 'Billing API', 'Core Platform', 'all', 'active', 'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBillingApiExampleKey', 'Primary transactional consumer for invoices and receipts.', '2026-04-29 08:10:00', '2026-05-01 10:35:00', '2026-05-05 00:00:00'),
('support-hub', 'Support Hub', 'Customer Care', 'email', 'paused', 'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAISupportHubExampleKey', 'Used for ticket replies and customer follow-ups.', '2026-04-21 13:25:00', NULL, '2026-05-05 00:00:00'),
('alerts-worker', 'Alerts Worker', 'Platform Automation', 'sms', 'active', 'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIAlertsWorkerExampleKey', 'Nightly alerting and incident notifications.', '2026-04-18 17:15:00', '2026-04-30 07:40:00', '2026-05-05 00:00:00');

INSERT INTO email_accounts (id, address, display_name, smtp_host, smtp_port, smtp_username, smtp_password, is_default, status, last_tested_at, created_at, updated_at) VALUES
('support', 'support@example.com', 'Support Desk', 'smtp.mail.example.com', 587, 'support@example.com', '••••••••••••', 1, 'healthy', '2026-05-04 09:20:00', '2026-04-20 09:00:00', '2026-05-05 00:00:00'),
('receipts', 'receipts@example.com', 'Receipts', 'smtp.send.example.com', 465, 'receipts@example.com', '••••••••••••', 0, 'warning', '2026-05-03 14:45:00', '2026-04-25 11:30:00', '2026-05-05 00:00:00');

INSERT INTO sms_credentials (id, username, password, status, last_synced_at, rotation_count, updated_at)
VALUES (1, 'api_user_id', '••••••••••••', 'connected', '2026-05-04 16:10:00', 2, '2026-05-05 00:00:00');

INSERT INTO messages (id, channel, service_id, sender, subject, content_mode, body, template_name, request_json, rendered_text, status, created_at, queued_at, accepted_at, provider_message_id, provider_response_json, error_code, error_message, cost, currency) VALUES
('msg-1001', 'email', 'billing-api', 'support@example.com', 'Invoice ready', 'plain', 'Your invoice is ready for download.', NULL, '{"channel":"email","serviceId":"billing-api","recipients":["finance@example.com"],"from":"support@example.com","subject":"Invoice ready","contentMode":"plain","body":"Your invoice is ready for download."}', 'Email accepted for delivery', 'accepted', '2026-05-04 15:20:00', '2026-05-04 15:20:00', '2026-05-04 15:20:00', NULL, NULL, NULL, NULL, NULL, NULL),
('msg-1002', 'sms', 'alerts-worker', 'AlertOps', NULL, 'template', 'Template payload for incident notification', 'incident-sms', '{"channel":"sms","serviceId":"alerts-worker","recipients":["+46700000000"],"senderName":"AlertOps","contentMode":"template","template":{"name":"incident-sms","data":{"incident":"INC-1001"}}}', 'SMS accepted for delivery', 'queued', '2026-05-04 18:40:00', '2026-05-04 18:40:00', NULL, NULL, NULL, NULL, NULL, NULL, NULL);

INSERT INTO message_recipients (message_id, ordinal, recipient, recipient_name, country_code) VALUES
('msg-1001', 0, 'finance@example.com', NULL, NULL),
('msg-1002', 0, '+46700000000', NULL, 'SE');

INSERT INTO activity_log (id, title, detail, tone, entity_type, entity_id, metadata_json, created_at) VALUES
('activity-1', 'Service key rotated', 'Billing API generated a new signing key.', 'success', 'service', 'billing-api', NULL, '2026-05-04 10:35:00'),
('activity-2', 'SMTP warning', 'Receipts account needs a reconnect check.', 'warning', 'email_account', 'receipts', NULL, '2026-05-03 14:45:00'),
('activity-3', 'SMS credentials synced', '46elks credentials were validated successfully.', 'info', 'sms_credentials', '1', NULL, '2026-05-04 16:10:00');

