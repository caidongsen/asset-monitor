package gmgop

import (
	proto "github.com/golang/protobuf/proto"
)

// Write proto message, length delimited
func MarshalMessage(msg proto.Message) ([]byte, error) {
	return proto.Marshal(msg)
}

// Read proto message, length delimited
func UnmarshalMessage(bz []byte, msg proto.Message) error {
	return proto.Unmarshal(bz, msg)
}

func (req *Request) CheckSign() error {
	sign := req.GetSign()
	req.Sign = nil
	data, err := MarshalMessage(req)
	if err != nil {
		return err
	}
	return CheckSign(data, req.Pubkey, sign)
}
