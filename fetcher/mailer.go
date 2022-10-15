package fetcher

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"log"
	"net/smtp"
	"strings"
	"time"
)

// Constants

const EmailMimeHeaders = "Content-Type: multipart/alternative; boundary=\"boundary-string\"" + CRLF +
	"MIME-Version: 1.0" + DblCrLf
const EmailSectionHeader = "--boundary-string" + CRLF + "Content-Type: %s; charset=\"utf-8\"" + CRLF +
	"Content-Transfer-Encoding: base64" + CRLF + "MIME-Version: 1.0" + DblCrLf
const DigestItemTextTemplate = "* %s - %s" + CRLF
const DigestItemHTMLTemplate = "<li><a href=\"%s\">%s</a></li>" + CRLF
const DigestHTMLTemplate = `<html>
<head>HackerNews Digest</head>
<body>
  <p>Hi!</p>
  <div>
  <ul>
  %s
  </ul>
  </div>
  <p>Generated: %s</p>
</body>
</html>%s`
const BoundaryString = "--boundary-string--"
const Base64LineLength = 76

// Mailer data type and its methods

type DigestMailer struct {
	smtpConfig SmtpConfig
}

func toBase64(input string) string {
	encoded := base64.StdEncoding.EncodeToString([]byte(input))
	normalized := ""
	inputLen := len(encoded)
	curIndex := 0

	for ; curIndex < len(encoded); curIndex += Base64LineLength {
		chunkSize := min(Base64LineLength, inputLen)
		normalized += fmt.Sprintf("%s"+CRLF, encoded[curIndex:curIndex+chunkSize])
		inputLen -= Base64LineLength
	}

	return normalized
}

// Prepare and send an email with the list of the provided news items
func (mailer *DigestMailer) SendEmail(digest *[]DigestItem, emailTo, emailSubject string) {
	if mailer.smtpConfig.Host == "" {
		log.Println("SMTP Host is empty. Skipping sending the Email")
		return
	}

	headers := map[string]string{
		"From":    mailer.smtpConfig.From,
		"Subject": emailSubject,
		"To":      emailTo,
		"Date":    time.Now().Format(time.RFC1123Z),
	}

	var messageBuilder, digestItemsHTMLBuilder, digestItemsTextBuilder strings.Builder

	for k, v := range headers {
		messageBuilder.WriteString(fmt.Sprintf("%s: %s"+CRLF, k, v))
	}

	digestItemsTextBuilder.WriteString("Hi!" + DblCrLf)

	for _, digestItem := range *digest {
		digestItemsHTMLBuilder.WriteString(fmt.Sprintf(DigestItemHTMLTemplate,
			digestItem.newsUrl, digestItem.newsTitle))
		digestItemsTextBuilder.WriteString(fmt.Sprintf(DigestItemTextTemplate, digestItem.newsTitle, digestItem.newsUrl))
	}

	messageBuilder.WriteString(EmailMimeHeaders)
	messageBuilder.WriteString(fmt.Sprintf(EmailSectionHeader, "text/plain"))
	messageBuilder.WriteString(toBase64(digestItemsTextBuilder.String()))
	messageBuilder.WriteString(CRLF)
	messageBuilder.WriteString(fmt.Sprintf(EmailSectionHeader, "text/html"))
	messageBuilder.WriteString(toBase64(fmt.Sprintf(DigestHTMLTemplate, digestItemsHTMLBuilder.String(),
		time.Now().Format(time.RFC1123Z), DblCrLf)))
	messageBuilder.WriteString(CRLF)
	messageBuilder.WriteString(BoundaryString)

	c, err := smtp.Dial(fmt.Sprintf("%s:%d", mailer.smtpConfig.Host, mailer.smtpConfig.Port))
	if err != nil {
		log.Fatal("EMAIL: ", err)
	}

	auth := smtp.PlainAuth("", mailer.smtpConfig.Username, mailer.smtpConfig.Password, mailer.smtpConfig.Host)

	if mailer.smtpConfig.UseTls {
		tlsconfig := &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         mailer.smtpConfig.Host,
		}
		_ = c.StartTLS(tlsconfig)
	}

	if err = c.Auth(auth); err != nil {
		log.Panic(err)
	}

	if err = c.Mail(mailer.smtpConfig.From); err != nil {
		log.Fatal("EMAIL_SENDER: ", err)
	}

	if err = c.Rcpt(emailTo); err != nil {
		log.Fatal("EMAIL_RECEIVER: ", err)
	}

	wc, err := c.Data()
	if err != nil {
		log.Fatal("EMAIL_START_CONTENT: ", err)
	}

	_, err = fmt.Fprint(wc, messageBuilder.String())
	if err != nil {
		log.Fatal("EMAIL_SET_CONTENT: ", err)
	}

	if err = wc.Close(); err != nil {
		log.Fatal("EMAIL_CLOSE_CONTENT: ", err)
	}

	if err = c.Quit(); err != nil {
		log.Fatal("EMAIL_QUIT: ", err)
	}
}
