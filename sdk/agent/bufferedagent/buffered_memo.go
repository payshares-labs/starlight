package bufferedagent

import (
	"encoding/base64"
	"fmt"
	"strings"

	xdr "github.com/stellar/go-xdr/xdr3"
)

type bufferedPaymentsMemo struct {
	ID       string
	Payments []BufferedPayment
}

func (m bufferedPaymentsMemo) String() string {
	sb := strings.Builder{}
	b64 := base64.NewEncoder(base64.StdEncoding, &sb)
	enc := xdr.NewEncoder(b64)
	_, err := enc.Encode(m)
	if err != nil {
		panic(fmt.Errorf("encoding buffered payments memo as json: %w", err))
	}
	return sb.String()
}

func parseBufferedPaymentMemo(memo string) (bufferedPaymentsMemo, error) {
	r := strings.NewReader(memo)
	b64 := base64.NewDecoder(base64.StdEncoding, r)
	dec := xdr.NewDecoder(b64)
	m := bufferedPaymentsMemo{}
	_, err := dec.Decode(&m)
	if err != nil {
		return bufferedPaymentsMemo{}, fmt.Errorf("decoding buffered payments memo from json: %w", err)
	}
	return m, nil
}
