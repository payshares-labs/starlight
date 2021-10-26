package bufferedagent

import (
	"encoding/json"
	"fmt"
	"strings"
)

type bufferedPaymentsMemo struct {
	ID       string
	Payments []BufferedPayment
}

func (m bufferedPaymentsMemo) String() string {
	sb := strings.Builder{}
	enc := json.NewEncoder(&sb)
	err := enc.Encode(m)
	if err != nil {
		panic(fmt.Errorf("encoding buffered payments memo as json: %w", err))
	}
	return sb.String()
}

func parseBufferedPaymentMemo(memo string) (bufferedPaymentsMemo, error) {
	r := strings.NewReader(memo)
	dec := json.NewDecoder(r)
	m := bufferedPaymentsMemo{}
	err := dec.Decode(&m)
	if err != nil {
		return bufferedPaymentsMemo{}, fmt.Errorf("decoding buffered payments memo from json: %w", err)
	}
	return m, nil
}
