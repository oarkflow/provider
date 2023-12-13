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

type GenericHttp struct {
	*protocol.HTTP
	AccessToken string
	ExpiresIn   int
}

func (r *GenericHttp) setupBasicAuth(auth *http.BasicAuth) error {
	var resp map[string]interface{}
	encoded := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", auth.Username, auth.Password)))
	data := make(map[string]interface{})
	if auth.Data != nil {
		for field, val := range auth.Data {
			data[field] = val
		}
	}
	headers := map[string]string{
		"Authorization": "Basic " + encoded,
	}
	if auth.Headers != nil {
		for key, val := range headers {
			headers[key] = val
		}
	}
	payload := protocol.Payload{
		URL:     auth.URL,
		Method:  auth.Method,
		Data:    data,
		Headers: headers,
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
	if auth.TokenField == "" {
		r.AccessToken = encoded
		r.Config.Headers["Authorization"] = "Bearer " + r.AccessToken
		return nil
	}
	if val, ok := resp[auth.TokenField]; ok {
		if r.Config.Headers == nil {
			r.Config.Headers = make(map[string]string)
		}
		r.AccessToken = val.(string)
		if auth.ExpiryField != "" {
			if v, ok := resp[auth.ExpiryField]; ok {
				r.ExpiresIn = int(v.(float64))
			}
		}
		r.Config.Headers["Authorization"] = "Bearer " + r.AccessToken
		return nil
	}
	return nil
}

func (r *GenericHttp) setupOAuth2(auth *http.OAuth2) error {
	var resp map[string]interface{}
	payload := protocol.Payload{
		URL:     auth.URL,
		Method:  "POST",
		Data:    auth.Data,
		Headers: auth.Headers,
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
	if auth.TokenField == "" {
		return errors.New("no token field defined")
	}
	if val, ok := resp[auth.TokenField]; ok {
		if r.Config.Headers == nil {
			r.Config.Headers = make(map[string]string)
		}
		r.AccessToken = val.(string)
		if auth.ExpiryField != "" {
			if v, ok := resp[auth.ExpiryField]; ok {
				r.ExpiresIn = int(v.(float64))
			}
		}
		
		r.Config.Headers["Authorization"] = "Bearer " + r.AccessToken
		return nil
	}
	return fmt.Errorf("invalid Credential: %s", string(bodyBytes))
}

func (r *GenericHttp) setupBearerToken(auth *http.BearerToken) error {
	if r.Config.Headers == nil {
		r.Config.Headers = make(map[string]string)
	}
	r.AccessToken = auth.Token
	r.Config.Headers["Authorization"] = "Bearer " + r.AccessToken
	return nil
}

func (r *GenericHttp) Setup() error {
	switch auth := r.HTTP.Config.Auth.(type) {
	case *http.BearerToken:
		return r.setupBearerToken(auth)
	case *http.BasicAuth:
		return r.setupBasicAuth(auth)
	case *http.OAuth2:
		return r.setupOAuth2(auth)
	default:
		return nil
	}
}

func (r *GenericHttp) SetService(provider protocol.Service) {
	r.HTTP = provider.(*protocol.HTTP)
}

type GenericSmpp struct {
	*protocol.SMPP
	AccessToken string
	ExpiresIn   int
}

func (r *GenericSmpp) SetService(provider protocol.Service) {
	r.SMPP = provider.(*protocol.SMPP)
}

type GenericSmtp struct {
	*protocol.SMTP
	AccessToken string
	ExpiresIn   int
}

func (r *GenericSmtp) SetService(provider protocol.Service) {
	r.SMTP = provider.(*protocol.SMTP)
}
