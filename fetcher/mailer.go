package fetcher

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"log"
	"net/smtp"
	"time"
)

// Constants

const EMAIL_MIME_HEADERS = "Content-Type: multipart/alternative; boundary=\"boundary-string\"" + CRLF +
	"MIME-Version: 1.0" + DBL_CRLF
const EMAIL_SECTION_HEADER = "--boundary-string" + CRLF + "Content-Type: %s; charset=\"utf-8\"" + CRLF +
	"Content-Transfer-Encoding: base64" + CRLF + "MIME-Version: 1.0" + DBL_CRLF
const DIGEST_ITEM_TEXT_TEMPLATE = "* %s - %s" + CRLF
const DIGEST_ITEM_HTML_TEMPLATE = "<li><a href=\"%s\">%s</a></li>" + CRLF
const DIGEST_HTML_TEMPLATE = `<html>
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
const BOUNDARY_STRING = "--boundary-string--"
const BASE64_LINE_LEN = 76

// Mailer data type and its methods

type DigestMailer struct {
	smtpConfig SmtpConfig
}

func toBase64(input string) string {
	encoded := base64.StdEncoding.EncodeToString([]byte(input))
	normalized := ""
	input_len := len(encoded)
	cur_index := 0
	for ; cur_index < len(encoded); cur_index += BASE64_LINE_LEN {
		chunk_size := min(BASE64_LINE_LEN, input_len)
		normalized += fmt.Sprintf("%s"+CRLF, encoded[cur_index:cur_index+chunk_size])
		input_len -= BASE64_LINE_LEN
	}
	return normalized
}

// Prepare and send an email with the list of the provided news items
func (mailer *DigestMailer) SendEmail(digest *[]DigestItem, emailTo string) {
	if mailer.smtpConfig.Host == "" {
		log.Println("SMTP Host is empty. Skipping sending the Email")
		return
	}

	headers := make(map[string]string)
	headers["From"] = mailer.smtpConfig.From
	headers["Subject"] = mailer.smtpConfig.Subject
	headers["To"] = emailTo
	headers["Date"] = time.Now().Format(time.RFC822)

	messageStart := ""
	for k, v := range headers {
		messageStart += fmt.Sprintf("%s: %s"+CRLF, k, v)
	}

	digestItemsHtml := ""
	digestItemsText := "Hi!" + DBL_CRLF
	for _, digestItem := range *digest {
		digestItemsHtml += fmt.Sprintf(DIGEST_ITEM_HTML_TEMPLATE,
			digestItem.newsUrl, digestItem.newsTitle)
		digestItemsText += fmt.Sprintf(DIGEST_ITEM_TEXT_TEMPLATE, digestItem.newsTitle, digestItem.newsUrl)
	}
	mime := EMAIL_MIME_HEADERS
	textHeader := fmt.Sprintf(EMAIL_SECTION_HEADER, "text/plain")
	htmlHeader := fmt.Sprintf(EMAIL_SECTION_HEADER, "text/html")
	digestHtml := fmt.Sprintf(DIGEST_HTML_TEMPLATE, digestItemsHtml, time.Now().Format(time.RFC1123Z), DBL_CRLF)

	msg := messageStart + mime +
		textHeader + toBase64(digestItemsText) + CRLF +
		htmlHeader + toBase64(digestHtml) + CRLF +
		BOUNDARY_STRING

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
		c.StartTLS(tlsconfig)
	}

	if err = c.Auth(auth); err != nil {
		log.Panic(err)
	}

	if err := c.Mail(mailer.smtpConfig.From); err != nil {
		log.Fatal("EMAIL_SENDER: ", err)
	}
	if err := c.Rcpt(emailTo); err != nil {
		log.Fatal("EMAIL_RECEIVER: ", err)
	}
	wc, err := c.Data()
	if err != nil {
		log.Fatal("EMAIL_START_CONTENT: ", err)
	}
	_, err = fmt.Fprint(wc, msg)
	if err != nil {
		log.Fatal("EMAIL_SET_CONTENT: ", err)
	}
	err = wc.Close()
	if err != nil {
		log.Fatal("EMAIL_CLOSE_CONTENT: ", err)
	}
	err = c.Quit()
	if err != nil {
		log.Fatal("EMAIL_QUIT: ", err)
	}
}
