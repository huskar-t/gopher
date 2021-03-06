package nats

import (
	"bytes"
	"github.com/huskar-t/gopher/infrastructure/json"
	"github.com/nats-io/nats.go"
	"strings"
)

// JsonEncoder is a JSON Encoder implementation for EncodedConn.
// This encoder will use the builtin encoding/json to Marshal
// and Unmarshal most types, including structs.
type JsonEncoder struct {
	// Empty
}

// Encode
func (je *JsonEncoder) Encode(subject string, v interface{}) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// Decode
func (je *JsonEncoder) Decode(subject string, data []byte, vPtr interface{}) (err error) {
	switch arg := vPtr.(type) {
	case *string:
		// If they want a string and it is a JSON string, strip quotes
		// This allows someone to send a struct but receive as a plain string
		// This cast should be efficient for Go 1.3 and beyond.
		str := string(data)
		if strings.HasPrefix(str, `"`) && strings.HasSuffix(str, `"`) {
			*arg = str[1 : len(str)-1]
		} else {
			*arg = str
		}
	case *[]byte:
		*arg = data
	default:
		//err = json.Unmarshal(data, arg)
		decoder := json.NewDecoder(bytes.NewReader(data))
		decoder.UseNumber()
		err = decoder.Decode(arg)
	}
	return
}

const JSON_ENCODER = "json2"

func init() {
	nats.RegisterEncoder(JSON_ENCODER, &JsonEncoder{})
}
