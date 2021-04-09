package json

import (
	"bytes"
	"encoding/json"

	"packetbeat/libbeat/common"
	"packetbeat/libbeat/logp"
	"packetbeat/libbeat/outputs"
)

type Encoder struct {
	Pretty bool
}

type Config struct {
	Pretty bool
}

var defaultConfig = Config{
	Pretty: false,
}

func init() {
	outputs.RegisterOutputCodec("json", func(cfg *common.Config) (outputs.Codec, error) {
		config := defaultConfig
		if cfg != nil {
			if err := cfg.Unpack(&config); err != nil {
				return nil, err
			}
		}

		return New(config.Pretty), nil
	})
}

func New(pretty bool) *Encoder {
	return &Encoder{pretty}
}

func (e *Encoder) Encode(event common.MapStr) ([]byte, error) {
	var err error
	var serializedEvent []byte

	if e.Pretty {
		serializedEvent, err = DisableEscapeHtmlMarshalIndent(event, "", "  ")
	} else {
		serializedEvent, err = DisableEscapeHtmlMarshal(event)
	}
	if err != nil {
		logp.Err("Fail to convert the event to JSON (%v): %#v", err, event)
	}

	return serializedEvent, err
}

func DisableEscapeHtmlMarshal(data interface{}) ([]byte, error) {
	bf := bytes.NewBuffer([]byte{})
	jsonEncoder := json.NewEncoder(bf)
	jsonEncoder.SetEscapeHTML(false)
	if err := jsonEncoder.Encode(data); err != nil {
		return []byte{}, err
	}
	return bf.Bytes(), nil
}

func DisableEscapeHtmlMarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	b, err := DisableEscapeHtmlMarshal(v)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	err = json.Indent(&buf, b, prefix, indent)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
