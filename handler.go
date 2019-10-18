package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"html/template"
	"log"
)

type Email struct {
	To       string `json:"to"`
	Subject  string `json:"subject"`
	Body     string `json:"body"`
	Code     string `json:"code"`
	Template string `json:"template"`
}

var templates = template.Must(template.ParseFiles("email-confirmation.html", "password-reset.html"))
var svc *ses.SES

func init() {
	if sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	}); err != nil {
		log.Printf("Failed to connect to AWS: %s", err.Error())
	} else {
		svc = ses.New(sess)
	}
}

func HandleRequest(r events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var b bytes.Buffer
	var e Email
	if err := json.Unmarshal([]byte(r.Body), &e); err != nil {
		return response(400, err.Error())
	} else if err := templates.ExecuteTemplate(bufio.NewWriter(&b), e.Template, e); err != nil {
		return response(400, err.Error())
	} else if _, err := svc.SendRawEmail(&ses.SendRawEmailInput{
		RawMessage: &ses.RawMessage{
			Data: []byte("Reply-To: noreply@hempconduit.com" +
				"\r\n" + "From: noreply@hempconduit.com" +
				"\r\n" + "To: " + e.To +
				"\r\n" + "Cc: " +
				"\r\n" + "Bcc: " +
				"\r\n" + "Subject: " + e.Subject +
				"\r\n" + "MIME-Version: 1.0" +
				"\r\n" + "Content-Type: text/html; charset=\"utf-8\"\r\n" +
				"\r\n" + b.String() + "\r\n"),
		},
	}); err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ses.ErrCodeMessageRejected:
				fmt.Println(ses.ErrCodeMessageRejected, aerr.Error())
			case ses.ErrCodeMailFromDomainNotVerifiedException:
				fmt.Println(ses.ErrCodeMailFromDomainNotVerifiedException, aerr.Error())
			case ses.ErrCodeConfigurationSetDoesNotExistException:
				fmt.Println(ses.ErrCodeConfigurationSetDoesNotExistException, aerr.Error())
			case ses.ErrCodeConfigurationSetSendingPausedException:
				fmt.Println(ses.ErrCodeConfigurationSetSendingPausedException, aerr.Error())
			case ses.ErrCodeAccountSendingPausedException:
				fmt.Println(ses.ErrCodeAccountSendingPausedException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
			return response(400, aerr.Error())
		} else {
			return response(500, err.Error())
		}
	} else {
		return response(200, `{ "success": "" }`)
	}
}

func response(code int, body string) (events.APIGatewayProxyResponse, error) {
	log.Printf(`response { code: %d, body: %s }`, code, body)
	return events.APIGatewayProxyResponse{
		StatusCode: code,
		Headers: map[string]string{
			"Access-Control-Allow-Origin": "*",
		},
		Body:            body,
		IsBase64Encoded: false,
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
