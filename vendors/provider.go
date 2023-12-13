package vendors

import (
	"github.com/oarkflow/protocol"
)

func NewGenericHttp() protocol.Service {
	return &GenericHttp{}
}

func NewGenericSmtp() protocol.Service {
	return &GenericSmtp{}
}

func NewGenericSmpp() protocol.Service {
	return &GenericSmpp{}
}

func NewRouteeHttpSms() protocol.Service {
	return &RouteeHttpSms{}
}

func NewSmstoHttpSms() protocol.Service {
	return &SmstoHttpSms{}
}
