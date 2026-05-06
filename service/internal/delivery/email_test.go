package delivery

import (
	"strings"
	"testing"
)

func TestEmailProvider_Send_SubjectEncoding(t *testing.T) {
	msg := buildEmailMessage(
		"test@example.com",
		[]string{"recipient@example.com"},
		"Test ÅÄÖ Subject",
		"Body content",
		"text/plain",
	)

	if strings.Contains(msg, "Subject: Test ÅÄÖ Subject") {
		t.Fatalf("subject was not encoded:\n%s", msg)
	}

	if !strings.Contains(msg, "Subject: =?utf-8?q?") {
		t.Fatalf("subject does not appear to be encoded correctly:\n%s", msg)
	}
}
