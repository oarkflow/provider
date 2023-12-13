package vendors

import (
	"encoding/json"
	"io"
	stdHttp "net/http"

	"github.com/oarkflow/protocol"
	"github.com/oarkflow/protocol/http"
)

type SmstoHttpSms struct {
	*protocol.HTTP
	AccessToken string
	ExpiresIn   int
}

func (r *SmstoHttpSms) setupOAuth2(auth *http.BasicAuth) error {
	var resp map[string]any
	payload := protocol.Payload{
		URL:    auth.URL,
		Method: "POST",
		Data: map[string]any{
			"client_id": auth.Username,
			"secret":    auth.Password,
		},
	}
	response, err := r.HTTP.Handle(payload)
	if err != nil {
		return err
	}
	bodyBytes, err := io.ReadAll(response.(*stdHttp.Response).Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(bodyBytes, &resp)
	if err != nil {
		return err
	}
	if val, ok := resp["jwt"]; ok {
		if r.Config.Headers == nil {
			r.Config.Headers = make(map[string]string)
		}
		r.AccessToken = val.(string)
		if resp["expires"] != nil {
			r.ExpiresIn = int(resp["expires"].(float64))
		}

		r.Config.Headers["Authorization"] = "Bearer " + r.AccessToken
	}
	return nil
}

func (r *SmstoHttpSms) setupBearerToken(auth *http.BearerToken) error {
	if r.Config.Headers == nil {
		r.Config.Headers = make(map[string]string)
	}
	r.AccessToken = auth.Token
	r.Config.Headers["Authorization"] = "Bearer " + r.AccessToken
	return nil
}

func (r *SmstoHttpSms) Setup() error {
	switch auth := r.HTTP.Config.Auth.(type) {
	case *http.BasicAuth:
		return r.setupOAuth2(auth)
	case *http.BearerToken:
		return r.setupBearerToken(auth)
	}
	return nil
}

func (r *SmstoHttpSms) SetService(provider protocol.Service) {
	r.HTTP = provider.(*protocol.HTTP)
}

func (r *SmstoHttpSms) Handle(payload protocol.Payload) (protocol.Response, error) {
	if payload.Data == nil {
		payload.Data = map[string]any{
			"sender_id":    payload.From,
			"to":           payload.To,
			"message":      payload.Message,
			"callback_url": payload.CallbackURL,
		}
	}
	return r.HTTP.Handle(payload)
}
