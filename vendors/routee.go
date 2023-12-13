package vendors

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	stdHttp "net/http"

	"github.com/oarkflow/protocol"
	"github.com/oarkflow/protocol/http"
)

type RouteeHttpSms struct {
	*protocol.HTTP
	AccessToken string
	ExpiresIn   int
}

func (r *RouteeHttpSms) setupBasicAuth(auth *http.BasicAuth) error {
	var resp map[string]any
	encoded := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", auth.Username, auth.Password)))
	payload := protocol.Payload{
		URL:    auth.URL,
		Method: "FORM",
		Data: map[string]any{
			"grant_type": "client_credentials",
		},
		Headers: map[string]string{
			"Authorization": "Basic " + encoded,
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
	if val, ok := resp["access_token"]; ok {
		if r.Config.Headers == nil {
			r.Config.Headers = make(map[string]string)
		}
		r.AccessToken = val.(string)
		r.ExpiresIn = int(resp["expires_in"].(float64))
		r.Config.Headers["Authorization"] = "Bearer " + r.AccessToken
		return nil
	}
	if val, ok := resp["developerMessage"]; ok {
		return errors.New(val.(string))
	}
	if val, ok := resp["error_description"]; ok {
		return errors.New(val.(string))
	}
	if _, ok := resp["error"]; ok {
		return errors.New(resp["message"].(string))
	}
	return nil
}

func (r *RouteeHttpSms) Setup() error {
	switch auth := r.HTTP.Config.Auth.(type) {
	case *http.BasicAuth:
		return r.setupBasicAuth(auth)
	}
	return nil
}

func (r *RouteeHttpSms) SetService(provider protocol.Service) {
	r.HTTP = provider.(*protocol.HTTP)
}

func (r *RouteeHttpSms) Handle(payload protocol.Payload) (protocol.Response, error) {
	if payload.Data == nil {
		payload.Data = map[string]any{
			"from": payload.From,
			"to":   payload.To,
			"body": payload.Message,
			"callback": map[string]any{
				"url":      payload.CallbackURL,
				"strategy": "OnCompletion",
			},
		}
	}
	return r.HTTP.Handle(payload)
}
