package provider

import (
	"github.com/oarkflow/protocol"
	
	"github.com/oarkflow/provider/vendors"
)

var Providers = map[string]func() protocol.Service{
	"generic-http-rest":  vendors.NewGenericHttp,
	"generic-http-sms":   vendors.NewGenericHttp,
	"generic-http-email": vendors.NewGenericHttp,
	"generic-smpp-sms":   vendors.NewGenericSmpp,
	"generic-smtp-email": vendors.NewGenericSmtp,
	"routee-http-sms":    vendors.NewRouteeHttpSms,
	"smsto-http-sms":     vendors.NewSmstoHttpSms,
}
