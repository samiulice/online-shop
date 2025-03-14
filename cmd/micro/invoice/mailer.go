package main

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"time"

	mail "github.com/xhit/go-simple-mail/v2"
)

//go:embed email-templates
var emailTemplateFS embed.FS

func (app *application) SendMail(from, to, subject, tmpl string, attachments []string, data interface{}) error {
	// HTML Email templates
	templateName := fmt.Sprintf("email-templates/%s.html.tmpl", tmpl)

	t, err := template.New("email-html").ParseFS(emailTemplateFS, templateName)
	if err != nil {
		app.errorLog.Println("1", err)
		return err
	}

	var tpl bytes.Buffer
	err = t.ExecuteTemplate(&tpl, "body", data)
	if err != nil {
		app.errorLog.Println(err)
		return err
	}

	formattedMessage := tpl.String()

	// //Plain text Email templates
	// templateName = fmt.Sprintf("email-templates/%s.plain.tmpl", tmpl)

	// t, err = template.New("email-plain").ParseFS(emailTemplateFS, templateName)
	// if err != nil {
	// 	app.errorLog.Println(err)
	// 	return err
	// }
	// err = t.ExecuteTemplate(&tpl, "body", data)
	// if err != nil {
	// 	app.errorLog.Println(err)
	// 	return err
	// }

	// plainMessage := tpl.String()

	//set up mail server
	server := mail.NewSMTPClient()
	server.Host = app.config.smtp.host
	server.Port = app.config.smtp.port
	server.Username = app.config.smtp.username
	server.Password = app.config.smtp.password
	server.Encryption = mail.EncryptionTLS
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	//Create smtp client for connecting to the smtp server
	smtpClient, err := server.Connect()
	if err != nil {
		return err
	}

	//send email
	email := mail.NewMSG()
	email.SetFrom(from).
		AddTo(to).
		SetSubject(subject)
	email.SetBody(mail.TextHTML, formattedMessage)
	// email.AddAlternative(mail.TextPlain, plainMessage)

	if len(attachments) > 0 {
		for _, path := range attachments{
			email.AddAttachment(path)
		}
	}
	err = email.Send(smtpClient)
	if err != nil {
		app.errorLog.Println(err)
		return err
	}
	app.infoLog.Println("Mail sent successfully")
	return nil
}
