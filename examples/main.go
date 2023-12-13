package main

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/oarkflow/protocol"
	"github.com/oarkflow/protocol/smpp"

	"github.com/oarkflow/provider"
)

var sms = protocol.Payload{
	From:    "00147932",
	To:      "99147932",
	Message: "Hello World!",
}

var email = protocol.Payload{
	To:      "s.baniya.np@gmail.com",
	From:    "s.baniya.np@gmail.com",
	Subject: "Test",
	Message: "<h1>This is test</h1",
}

var httpPayload = protocol.Payload{
	URL:    "http://localhost:3000",
	Method: "POST",
	Data: map[string]any{
		"id":         1,
		"first_name": "Sujit",
		"last_name":  "Baniya",
	},
}

var smsPayload = protocol.Payload{
	Data: map[string]any{
		"sender_id": "SMSto",
		"to":        "+9779856034616",
		"message":   "This is test",
	},
}

func main() {
	smtpEmailTest()
	// httpSmsTest()
	// httpSmsClientTest()
	// smppTest()
}

func httpSmsTest() {
	prov := provider.ServiceProvider{
		Host:    "https://api.sms.to/sms/estimate",
		Method:  "POST",
		AuthUrl: "https://auth.sms.to/oauth/token",
		AuthData: map[string]any{
			"client_id": "dfdsf",
			"secret":    "sdfds",
		},
		ServiceType: "http",
		AuthType:    "oauth2",
		Service:     "rest",
		Name:        "Test HTTP",
		Slug:        "test-http",
		TokenField:  "jwt",
	}
	response, err := prov.Handle(smsPayload)
	if err != nil {
		panic(err)
	}
	bodyBytes, err := io.ReadAll(response.(*http.Response).Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(bodyBytes))
}

func httpSmsClientTest() {
	prov := provider.ServiceProvider{
		ServiceType: "http",
		AuthUrl:     "https://msas.local.intergo.co/oauth/token",
		Username:    "9fKgec0vCE4rJZOD",
		Password:    "xmdqzFT6Cl6NxLgjTFwbFoxNOO04XcXm",
		Service:     "sms",
		Name:        "SMSto",
		Slug:        "smsto",
		TokenField:  "jwt",
	}
	for i := 0; i < 10; i++ {
		response, err := prov.Handle(smsPayload)
		if err != nil {
			panic(err)
		}
		bodyBytes, err := io.ReadAll(response.(*http.Response).Body)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(bodyBytes))
	}
}

func smtpEmailTest() {
	prov := provider.ServiceProvider{
		Name:        "Localhost SMTP",
		Slug:        "localhost-smtp",
		Host:        "localhost",
		Username:    "",
		Password:    "",
		Encryption:  "tls",
		ServiceType: "smtp",
		Service:     "email",
		FromAddress: "s.baniya.np@gmail.com",
		FromName:    "Sujit Baniya",
		Port:        1025,
	}
	response, err := prov.Handle(email)
	if err != nil {
		panic(err)
	}
	fmt.Println(response)
}

func smppTest() {
	onMessageReport := func(manager *smpp.Manager, sms *smpp.Message, parts []*smpp.Part) {
		fmt.Println("Message Report", sms)
		for _, part := range parts {
			fmt.Println(part)
		}
	}
	prov := provider.ServiceProvider{
		Name:            "Test-Smpp",
		Slug:            "test-smpp",
		Host:            "localhost",
		Port:            2775,
		Username:        "147932",
		Password:        "a16fb9",
		ServiceType:     "smpp",
		Service:         "sms",
		ReadTimeout:     10 * time.Second,
		WriteTimeout:    10 * time.Second,
		EnquiryInterval: time.Minute,
		EnquiryTimeout:  time.Minute,
		MaxConnection:   2,
		Throttle:        100,
	}
	prov.Handle(sms, onMessageReport)
	time.Sleep(20 * time.Second)
}
