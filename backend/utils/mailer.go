package utils

import (
	"fmt"
	"os"
	"gopkg.in/mail.v2"
)

func SendEmail(to string, code string) error {
	smtpUser := os.Getenv("SMTP_ACCOUNT")
	smtpPass := os.Getenv("SMTP_PASSWORD")

	m := mail.NewMessage()
	m.SetHeader("From", smtpUser)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "【数据采集脚本管理系统】验证码通知")
	m.SetBody("text/plain", fmt.Sprintf("您好，您的验证码为：%s，5分钟内有效，请勿泄露。", code))

	d := mail.NewDialer("smtp.qq.com", 465, smtpUser, smtpPass) //587

	return d.DialAndSend(m)
}
