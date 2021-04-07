package email

import (
	"errors"
	"fmt"
	"gopkg.in/gomail.v2"
)

type SMTPSetting struct {
	ID          string `json:"id,omitempty"`
	SiteID      string `json:"siteID" gorm:"unique_index"` // 关联的项目
	Enabled     bool   `json:"enabled"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	SSL         bool   `json:"ssl"`
	FromAddress string `json:"fromAddress"`
	FromName    string `json:"fromName"`
}

func sendMail(settings *SMTPSetting, to string, subject, body, mailType string) error {
	// fmt.Println(conf)
	if !settings.Enabled {
		return errors.New("SMTP not Enabled")
	}

	if settings.Port <= 0 || settings.Port > 65535 {
		return errors.New("SMTP port invalid")
	}

	if settings.Username == "" {
		return errors.New("SMTP username empty")
	}

	if settings.Password == "" {
		return errors.New("SMTP password empty")
	}

	if settings.FromAddress == "" {
		return errors.New("SMTP from address empty")
	}

	d := gomail.NewDialer(settings.Host, settings.Port, settings.Username, settings.Password)

	m := gomail.NewMessage()

	if settings.FromName == "" {
		m.SetHeader("From", settings.FromAddress)
	} else {
		m.SetAddressHeader("From", settings.FromAddress, settings.FromName)
	}
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)

	contentType := fmt.Sprintf("text/%s", mailType)
	m.SetBody(contentType, body)

	// Send the email to Bob, Cora and Dan.
	if err := d.DialAndSend(m); err != nil {
		return err
	}
	return nil
}
