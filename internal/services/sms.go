package services

import (
	"github.com/AfricasTalkingLtd/africastalking-go/sms"
	"log"
	"strings"
)

type SMSServiceInterface interface {
	Send(recipient, message string)
}

type SMSService struct {
	client   sms.Service
	senderID string
}

func NewSMSService(username, apiKey, env string) SMSServiceInterface {
	client := sms.NewService(username, apiKey, env)

	var senderID string
	if env == "sandbox" {
		senderID = "AFRICASTKNG"
	}

	return &SMSService{client: client, senderID: senderID}
}

func (s *SMSService) Send(recipient, message string) {
	if s.client.Username == "" {
		log.Println("SMS service not initialized, skipping send.")
		return
	}

	recipients := []string{recipient}
	to := strings.Join(recipients, ",")

	_, err := s.client.Send(to, message, s.senderID)

	if err != nil {
		log.Printf("Failed to send SMS to %s: %v", recipient, err)
	} else {
		log.Printf("SMS sent successfully to %s", recipient)
	}
}
