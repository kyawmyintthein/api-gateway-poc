package bodymodifier

import (
	"log"

	"github.com/google/martian/parse"
	"github.com/kyawmyintthein/api-gateway-poc/plugins/bodymodifier/modifier"
)

func init() {
	parse.Register("body.CustomModifier", FromJSON)
}

func FromJSON(b []byte) (*parse.Result, error) {
	log.Println("body.Modifier Request received")
	msg, err := modifier.FromJSON(b)
	if err != nil {
		return nil, err
	}

	return parse.NewResult(msg, msg.Scope)
}
