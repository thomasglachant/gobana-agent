package core

import (
	"crypto/tls"
	"fmt"
	gomail "gopkg.in/mail.v2"
)

func SendEmail(smtp *SMTPConfig, to, subject, template string, vars map[string]interface{}) error {
	if !CheckEmailTemplateExists(template) {
		return fmt.Errorf("email template %s does not exist", template)
	}

	vars["SUBJECT"] = subject

	m := gomail.NewMessage()

	m.SetAddressHeader("From", smtp.FromEmail, smtp.FromName)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", TemplateToHTML([]string{
		fmt.Sprintf("templates/email/%s.txt.tmpl", template),
		"templates/email/_layout.txt.tmpl",
	}, vars))
	m.SetBody("text/html", TemplateToString([]string{
		fmt.Sprintf("templates/email/%s.gohtml", template),
		"templates/email/_layout.gohtml",
	}, vars))
	if imgLogo, imgErr := AssetAsReader("assets/email/spooter-light.png"); imgErr == nil {
		m.EmbedReader("logo.png", imgLogo)
	}
	d := gomail.NewDialer(smtp.Host, smtp.Port, smtp.Username, smtp.Password)
	d.TLSConfig = &tls.Config{
		ServerName:         smtp.Host,
		InsecureSkipVerify: !smtp.SSLEnabled, //nolint:gosec
	}

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("unable to send email : %s", err)
	}

	return nil
}

func CheckEmailTemplateExists(template string) bool {
	if !CheckTemplateExists(fmt.Sprintf("templates/email/%s.txt.tmpl", template)) {
		return false
	}
	if !CheckTemplateExists(fmt.Sprintf("templates/email/%s.gohtml", template)) {
		return false
	}
	return true
}
