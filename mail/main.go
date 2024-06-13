package mail

import (
	"fmt"
	"gopkg.in/gomail.v2"
)

// SendEmail 发送邮箱验证码
// @param to 收件人
// @param subject 主题
// @param body 邮件内容
func SendEmail(to, subject, body string) (err error) {
	m := gomail.NewMessage()
	m.SetHeader("From", "noreply@service.fragpunk.com")
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)
	d := gomail.NewDialer("smtpdm-ap-southeast-1.aliyuncs.com",
		465,
		"noreply@service.fragpunk.com",
		"Q4t0FTfj7KFt1Wcm")
	// Send the email to Bob, Cora and Dan.
	if err := d.DialAndSend(m); err != nil {
		fmt.Println(err)
	}
	fmt.Println("success")
	return
}
