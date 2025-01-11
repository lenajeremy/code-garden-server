package emails

import (
	"code-garden-server/config"
	"fmt"
	"log"

	"github.com/resend/resend-go/v2"
)

type Mail struct {
	Emails  []string
	Html    string
	Text    string
	Subject string
}

func SendMail(v Mail) error {
	apiKey := config.GetEnv("RESEND_API_KEY")

	client := resend.NewClient(apiKey)

	params := &resend.SendEmailRequest{
		From:    "Jeremiah <test@craftmycv.xyz>",
		To:      v.Emails,
		Html:    v.Html,
		Text:    v.Text,
		Subject: v.Subject,
	}

	sent, err := client.Emails.Send(params)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	log.Println(sent.Id)
	return nil
}
