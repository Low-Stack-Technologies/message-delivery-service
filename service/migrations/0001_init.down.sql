-- +migrate Down
DROP TABLE IF EXISTS activity_log;
DROP TABLE IF EXISTS message_recipients;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS sms_credentials;
DROP TABLE IF EXISTS email_accounts;
DROP TABLE IF EXISTS services;
DROP TABLE IF EXISTS app_settings;

