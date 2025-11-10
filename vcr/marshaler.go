package vcr

import (
	"bytes"
	"fmt"

	"go.yaml.in/yaml/v4"
)

func customMarshaler(in any) ([]byte, error) {
	var buff bytes.Buffer
	enc := yaml.NewEncoder(&buff)
	enc.SetIndent(2)
	enc.CompactSeqIndent()
	if err := enc.Encode(in); err != nil {
		return nil, fmt.Errorf("vcr: unable to encode to yaml: %w", err)
	}
	return buff.Bytes(), nil
}
