// Package modifier exposes a request modifier for generating bodies
// from the querystring params
package modifier

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"

	kazaam "gopkg.in/qntfy/kazaam.v3"
)

type Config struct {
	URLPattern  string `json:"url_pattern"`
	Template    string `json:"template"`
	Method      string `json:"method"`
	ContentType string `json:"content_type"`
}

type Payload2BodyModifier struct {
	template    string
	method      string
	contentType string
}

func (m *Payload2BodyModifier) ModifyRequest(req *http.Request) error {
	if req.Body == nil {
		return nil
	}
	payloadBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}
	req.Body.Close()
	k, err := kazaam.NewKazaam(m.template)
	if err != nil {
		return err
	}

	tranformedDataBytes, err := k.TransformJSONString(string(payloadBytes))
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	buf.Read(tranformedDataBytes)
	if m.method != "" {
		req.Method = m.method
	}
	if m.contentType != "" {
		req.Header.Set("Content-Type", m.contentType)
	} else {
		req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	}
	req.ContentLength = int64(len(tranformedDataBytes))
	req.Body = ioutil.NopCloser(bytes.NewReader(tranformedDataBytes))

	return nil
}

func FromJSON(b []byte) (*Payload2BodyModifier, error) {
	cfg := &Config{}
	if err := json.Unmarshal(b, cfg); err != nil {
		return nil, err
	}

	bytes, err := b64.StdEncoding.DecodeString(cfg.Template) // Converting data
	if err != nil {
		return nil, err
	}

	return &Payload2BodyModifier{
		template: string(bytes),
		method:   cfg.Method,
	}, nil
}
