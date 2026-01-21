package email

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/gomail.v2"
	"github.com/ieasydevops/demo-scrapy/internal/models"
)

func SendEmail(to string, announcements []models.Announcement) error {
	if len(announcements) == 0 {
		return nil
	}

	var content strings.Builder
	content.WriteString("<h2>今日新增公告</h2>")
	content.WriteString("<ul>")
	for _, ann := range announcements {
		content.WriteString(fmt.Sprintf("<li><a href='%s'>%s</a> - %s</li>", ann.URL, ann.Title, ann.PublishDate))
	}
	content.WriteString("</ul>")

	m := gomail.NewMessage()
	
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := 587
	
	if smtpHost == "" {
		smtpHost = "smtp.qq.com"
	}
	if smtpUser == "" {
		smtpUser = "403608355@qq.com"
	}
	
	m.SetHeader("From", smtpUser)
	m.SetHeader("To", to)
	m.SetHeader("Subject", fmt.Sprintf("政府采购网公告通知 - %d条新公告", len(announcements)))
	m.SetBody("text/html", content.String())

	d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)

	return d.DialAndSend(m)
}
