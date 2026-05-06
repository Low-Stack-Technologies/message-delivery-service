-- +migrate Down
DELETE FROM activity_log;
DELETE FROM message_recipients;
DELETE FROM messages;
DELETE FROM sms_credentials;
DELETE FROM email_accounts;
DELETE FROM services;
DELETE FROM app_settings;

