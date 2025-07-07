package fetcher

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBodyToBase64(t *testing.T) {
	body := "Some message that should be encoded to base64 and splitted to lines each shorter than 80 symbols"
	expected := "U29tZSBtZXNzYWdlIHRoYXQgc2hvdWxkIGJlIGVuY29kZWQgdG8gYmFzZTY0IGFuZCBzcGxpdHRl" + CRLF +
		"ZCB0byBsaW5lcyBlYWNoIHNob3J0ZXIgdGhhbiA4MCBzeW1ib2xz" + CRLF
	encoded := toBase64(body)

	if expected != encoded {
		t.Fatalf("\n%s\nIS NOT equal to \n%s\n", encoded, expected)
	}
}

func TestSendMail(t *testing.T) {
	mailer := DigestMailer{}
	assert.NotPanics(t, func() {
		mailer.SendEmail(&[]DigestItem{
			{id: 1, newsTitle: "t", newsUrl: "url", createdAt: 12312}}, "", "")
	}, "SendEmail should not panic with empty parameters")
}
