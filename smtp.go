package sqlconf

import (
	"crypto/tls"
	"errors"
	"fmt"

	"log"
	"net/smtp"

	"os"
	"strings"
)

// MailSetup ...
type Mail struct {
	Host     string
	Port     string
	Username string
	Password string

	From    string
	To      []string
	Cc      []string
	Bcc     []string
	Body    string
	Subject string

	Message string
}

// LoginAuth for starttls
type LoginAuth struct {
	username string
	password string
}

// NewLoginAuth required for starttls
func NewLoginAuth(username, password string) smtp.Auth {
	return &LoginAuth{username, password}
}

// Start required for starttls
func (a *LoginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte{}, nil
}

// Next required for starttls
func (a *LoginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		default:
			return nil, errors.New("Unknown fromServer")
		}
	}
	return nil, nil
}

func (m *Mail) WithSMTPEnv(envHost, envPort, envUsername, envPassword string) *Mail {
	if envHost == "" || envPort == "" {
		log.Fatal("please set your env variables first. ie.: export " + envHost + "=smtp.office365.com")
	}

	m.Host = os.Getenv(envHost)
	m.Port = os.Getenv(envPort)

	m.Username = os.Getenv(envUsername)
	m.Password = os.Getenv(envPassword)

	return m
}

func (m *Mail) WithMailEnv(envFrom, envTo, envCC, envBCC string) *Mail {
	m.From = os.Getenv(envFrom)
	m.To = StringToSlice(os.Getenv(envTo))
	m.Cc = StringToSlice(os.Getenv(envCC))
	m.Bcc = StringToSlice(os.Getenv(envBCC))

	return m
}

func (m *Mail) WithMessage(subject, body string) *Mail {

	m.Subject = subject
	m.Body = body

	headers := make(map[string]string)
	headers["From"] = m.From
	headers["To"] = strings.Join(m.To, ";")
	if len(m.Cc) > 0 {
		headers["Cc"] = strings.Join(m.Cc, ";")
	}
	if len(m.Bcc) > 0 {
		headers["Bcc"] = strings.Join(m.Bcc, ";")
	}

	headers["Subject"] = m.Subject
	headers["Content-Type"] = "text/html"

	msg := ""
	for k, v := range headers {
		msg += fmt.Sprintf("%s: %s\r\n", k, v)
	}

	msg += "\r\n" + m.Body

	m.Message = msg

	return m
}

func (m *Mail) SendMail() error {
	hostPort := strings.Join([]string{m.Host, m.Port}, ":")

	var err error
	if m.Username == "" && m.Password == "" {
		log.Println("smtp using Anonymous ...")
		err = smtp.SendMail(hostPort, nil, m.From, m.To, []byte(m.Message))
	} else {
		log.Println("using PlainAuth ...")
		auth := smtp.PlainAuth("", m.Username, m.Password, m.Host)
		err = smtp.SendMail(hostPort, auth, m.From, m.To, []byte(m.Message))
	}

	if err != nil {
		log.Println(err)
		return err
	}

	log.Println("send mail: OK")
	return nil
}

// SendMailStartTLS ...
func (m *Mail) SendMailStartTLS() error {
	log.Println("using STARTTLS ...")

	hostPort := strings.Join([]string{m.Host, m.Port}, ":")

	smtpClient, err := smtp.Dial(hostPort)
	if err != nil {
		log.Println(hostPort, err)
		return err
	}
	defer smtpClient.Close()

	if ok, _ := smtpClient.Extension("STARTTLS"); ok {
		cfg := &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         m.Host}
		if err := smtpClient.StartTLS(cfg); err != nil {
			log.Println(err)
			return err
		}
	}

	a := NewLoginAuth(m.Username, m.Password)
	if ok, _ := smtpClient.Extension("AUTH"); ok {
		if err := smtpClient.Auth(a); err != nil {
			log.Println(err)
			return err
		}
	}

	if err := smtpClient.Mail(m.From); err != nil {
		log.Println(err)
		return err
	}

	for _, addr := range m.To {
		if strings.Index(addr, "@") < 0 {
			continue
		}
		if err := smtpClient.Rcpt(addr); err != nil {
			log.Println(err)
			return err
		}
	}

	w, err := smtpClient.Data()
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = w.Write([]byte(m.Message))
	if err != nil {
		log.Println(err)
		return err
	}

	err = w.Close()
	if err != nil {
		log.Println(err)
		return err
	}

	smtpClient.Quit()

	log.Println("send mail: OK")
	return nil
}
