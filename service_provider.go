package provider

import (
	"errors"
	"fmt"
	"time"
	
	"github.com/oarkflow/frame/server/render"
	"github.com/oarkflow/protocol"
	"github.com/oarkflow/protocol/http"
	"github.com/oarkflow/protocol/smpp"
	"github.com/oarkflow/protocol/smpp/pdu"
	"github.com/oarkflow/protocol/smtp"
)

type Provider struct {
	Name             string                 `json:"name"`
	Slug             string                 `json:"slug"`
	Host             string                 `json:"host"`
	Method           string                 `json:"method"`
	Port             int                    `json:"port"`
	Username         string                 `json:"username"`
	Password         string                 `json:"password"`
	Token            string                 `json:"token"`
	AuthType         string                 `json:"auth_type"`
	AuthUrl          string                 `json:"auth_url"`
	ClientID         string                 `json:"client_id"`
	Secret           string                 `json:"secret"`
	GrantType        string                 `json:"grant_type"`
	TokenField       string                 `json:"token_field"`
	ExpiryField      string                 `json:"expiry_field"`
	Encryption       string                 `json:"encryption"`
	FromAddress      string                 `json:"from_address"`
	SystemType       string                 `json:"system_type"`
	ServiceType      protocol.Type          `json:"service_type"`
	Service          string                 `json:"service"`
	FromName         string                 `json:"from_name"`
	ReadTimeout      time.Duration          `json:"read_timeout"`
	RequestTimeout   time.Duration          `json:"request_timeout"`
	WriteTimeout     time.Duration          `json:"write_timeout"`
	EnquiryInterval  time.Duration          `json:"enquiry_interval"`
	EnquiryTimeout   time.Duration          `json:"enquiry_timeout"`
	MaxConnection    int                    `json:"max_connection"`
	Throttle         int                    `json:"throttle"`
	UseAllConnection bool                   `json:"use_all_connection"`
	AutoRebind       bool                   `json:"auto_rebind"`
	RetryWaitMin     time.Duration          `json:"retry_wait_min"`
	RetryWaitMax     time.Duration          `json:"retry_wait_max"`
	RetryMax         int                    `json:"retry_max"`
	RespReadLimit    int64                  `json:"resp_read_limit"`
	KillIdleConn     bool                   `json:"kill_idle_conn"`
	MaxPoolSize      int                    `json:"max_pool_size"`
	Headers          map[string]string      `json:"headers"`
	AuthData         map[string]interface{} `json:"auth_data"`
	AuthHeaders      map[string]string      `json:"auth_headers"`
	HtmlEngine       *render.HtmlEngine
	HandlePDU        func(pdu pdu.Body)
	service          protocol.Service
}

func (provider *Provider) GetService(messageHandler ...any) (protocol.Service, error) {
	service, err := NewServiceProvider(provider, messageHandler...)
	if err != nil {
		return nil, err
	}
	key := fmt.Sprintf("%s-%s-%s", provider.Slug, provider.ServiceType, provider.Service)
	if function, ok := Providers[key]; ok {
		serviceProvider := function()
		serviceProvider.SetService(service)
		return serviceProvider, nil
	}
	key = fmt.Sprintf("%s-%s-%s", "generic", provider.ServiceType, provider.Service)
	if function, ok := Providers[key]; ok {
		serviceProvider := function()
		serviceProvider.SetService(service)
		return serviceProvider, nil
	}
	return nil, errors.New("no service provider")
}

func (provider *Provider) Handle(payload protocol.Payload, messageHandler ...any) (protocol.Response, error) {
	if provider.service == nil {
		service, err := provider.GetService(messageHandler...)
		if err != nil {
			return nil, err
		}
		err = service.Setup()
		if err != nil {
			return nil, err
		}
		provider.service = service
	}
	return provider.service.Handle(payload)
}

func (provider *Provider) Setup(messageHandler ...any) error {
	if provider.service == nil {
		service, err := provider.GetService(messageHandler...)
		if err != nil {
			return err
		}
		return service.Setup()
	}
	return nil
}

func NewServiceProvider(provider *Provider, messageHandler ...any) (protocol.Service, error) {
	switch provider.ServiceType {
	case protocol.Smtp:
		if provider.Host == "" || provider.Port == 0 {
			return nil, errors.New("no host detail")
		}
		return protocol.NewSMTP(smtp.Config{
			Host:        provider.Host,
			Username:    provider.Username,
			Password:    provider.Password,
			Encryption:  provider.Encryption,
			FromAddress: provider.FromAddress,
			FromName:    provider.FromName,
			Port:        provider.Port,
		}, provider.HtmlEngine)
	case protocol.Smpp:
		if provider.Host == "" || provider.Port == 0 {
			return nil, errors.New("no host detail")
		}
		if provider.Username == "" || provider.Password == "" {
			return nil, errors.New("no auth detail")
		}
		if provider.ReadTimeout == 0 {
			provider.ReadTimeout = 10
		}
		if provider.WriteTimeout == 0 {
			provider.WriteTimeout = 10
		}
		if provider.EnquiryInterval == 0 {
			provider.EnquiryInterval = 20
		}
		if provider.EnquiryTimeout == 0 {
			provider.EnquiryTimeout = 50
		}
		if provider.MaxConnection == 0 {
			provider.MaxConnection = 1
		}
		onMessageReport := func(manager *smpp.Manager, sms *smpp.Message, parts []*smpp.Part) {
			fmt.Println("Message Report", sms)
		}
		if len(messageHandler) > 0 {
			switch funcs := messageHandler[0].(type) {
			case func(manager *smpp.Manager, sms *smpp.Message, parts []*smpp.Part):
				onMessageReport = funcs
			}
		}
		return protocol.NewSMPP(smpp.Setting{
			Name: provider.Name,
			Slug: provider.Slug,
			URL:  fmt.Sprintf("%s:%d", provider.Host, provider.Port),
			Auth: smpp.Auth{
				SystemID:   provider.Username,
				Password:   provider.Password,
				SystemType: provider.SystemType,
			},
			ReadTimeout:      provider.ReadTimeout * time.Second,
			WriteTimeout:     provider.WriteTimeout * time.Second,
			EnquiryInterval:  provider.EnquiryInterval * time.Second,
			EnquiryTimeout:   provider.EnquiryTimeout * time.Second,
			MaxConnection:    provider.MaxConnection,
			Throttle:         provider.Throttle,
			UseAllConnection: provider.UseAllConnection,
			HandlePDU:        provider.HandlePDU,
			AutoRebind:       provider.AutoRebind,
			OnMessageReport:  onMessageReport,
		})
	case protocol.Http:
		var auth http.Auth
		if provider.AuthType == "oauth2" {
			auth = &http.OAuth2{
				URL:         provider.AuthUrl,
				ClientID:    provider.ClientID,
				Secret:      provider.Secret,
				GrantType:   provider.GrantType,
				Data:        provider.AuthData,
				Headers:     provider.AuthHeaders,
				TokenField:  provider.TokenField,
				ExpiryField: provider.ExpiryField,
			}
		} else if provider.AuthType == "api-key" {
			auth = &http.BearerToken{
				URL:   provider.AuthUrl,
				Token: provider.Token,
			}
		} else if provider.AuthType == "basic" {
			auth = &http.BasicAuth{
				URL:         provider.AuthUrl,
				Username:    provider.Username,
				Password:    provider.Password,
				Data:        provider.AuthData,
				Headers:     provider.AuthHeaders,
				TokenField:  provider.TokenField,
				ExpiryField: provider.ExpiryField,
			}
		}
		if provider.RetryWaitMin == 0 {
			provider.RetryWaitMin = 10
		}
		if provider.RetryWaitMax == 0 {
			provider.RetryWaitMax = 15
		}
		if provider.RequestTimeout == 0 {
			provider.RequestTimeout = 15
		}
		return protocol.NewHTTP(&http.Options{
			URL:           provider.Host,
			Method:        provider.Method,
			RetryWaitMin:  provider.RetryWaitMin * time.Second,
			RetryWaitMax:  provider.RetryWaitMax * time.Second,
			Timeout:       provider.RequestTimeout * time.Second,
			RetryMax:      provider.RetryMax,
			RespReadLimit: provider.RespReadLimit,
			KillIdleConn:  provider.KillIdleConn,
			MaxPoolSize:   provider.MaxPoolSize,
			ReqPerSec:     provider.Throttle,
			Headers:       provider.Headers,
			Auth:          auth,
		})
	default:
		return nil, errors.New("invalid service ty[e")
	}
}
