package string

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/own-home/central/pkg/payload"
)

// Handler implements `payload.Handler` and is capable of extracting data
// from strings
type Handler struct{}

// Parse parses the given payload and returns the extracted value. It implements
// the `Parse()` method of `payload.Handler`
func (h Handler) Parse(body []byte, cfg payload.HandlerSpec) (interface{}, error) {
	p := string(body)
	if reg, hasRegex := cfg["regex"]; hasRegex {
		r, ok := reg.(string)
		if !ok {
			return nil, fmt.Errorf("regex argument must be a string")
		}

		re, err := regexp.Compile(r)
		if err != nil {
			return nil, err
		}

		cptGrp, _, cIsValid := cfg.GetInt("group")
		grpIdx, _, iIsValid := cfg.GetInt("index")

		useGroup := cIsValid && iIsValid
		useAll := iIsValid

		if useGroup {
			matches := re.FindAllStringSubmatch(p, -1)

			if len(matches) > cptGrp {
				if len(matches[cptGrp]) > grpIdx {
					return matches[cptGrp][grpIdx], nil
				}

				return nil, errors.New("index out of bounds")
			}

			return nil, errors.New("capture group out of bounds")

		} else if useAll {
			matches := re.FindAllString(p, -1)

			if len(matches) > grpIdx {
				return matches[grpIdx], nil
			}

			return nil, errors.New("indexs out of bounds")
		}

		return re.FindString(p), nil
	}

	return p, nil
}

func init() {
	payload.MustRegisterType("string", Handler{})
}
