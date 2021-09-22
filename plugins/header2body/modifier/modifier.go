// Package modifier exposes a request modifier for generating bodies
// from the querystring params
package modifier

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"text/template"
)

type Config struct {
	KeysToExtract []string `json:"keys_to_extract"`
	Template      string   `json:"template"`
	Method        string   `json:"method"`
	ContentType   string   `json:"content_type"`
}

type Header2BodyModifier struct {
	keysToExtract []string
	template      *template.Template
	method        string
	contentType   string
}

func (m *Header2BodyModifier) ModifyRequest(req *http.Request) error {
	buf := new(bytes.Buffer)
	if err := m.template.Execute(buf, req.Header); err != nil {
		return err
	}

	if m.method != "" {
		req.Method = m.method
	}
	if m.contentType != "" {
		req.Header.Set("Content-Type", m.contentType)
	} else {
		req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	}
	req.ContentLength = int64(buf.Len())
	req.Body = ioutil.NopCloser(buf)

	return nil
}

func FromJSON(b []byte) (*Header2BodyModifier, error) {
	cfg := &Config{}
	if err := json.Unmarshal(b, cfg); err != nil {
		return nil, err
	}

	tmpl, err := template.New("header2body_modifier").Parse(cfg.Template)
	if err != nil {
		return nil, err
	}

	return &Header2BodyModifier{
		keysToExtract: cfg.KeysToExtract,
		template:      tmpl,
		method:        cfg.Method,
	}, nil
}
