package querystringtobody

import (
	"github.com/google/martian/parse"
	"github.com/kyawmyintthein/api-gateway-poc/plugins/header2body/modifier"
)

func init() {
	parse.Register("body.FromHeader", FromJSON)
}

func FromJSON(b []byte) (*parse.Result, error) {
	msg, err := modifier.FromJSON(b)
	if err != nil {
		return nil, err
	}

	return parse.NewResult(msg, []parse.ModifierType{parse.Request})
}
