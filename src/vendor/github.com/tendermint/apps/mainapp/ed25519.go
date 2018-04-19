package mainapp

import "errors"
import "github.com/tendermint/ed25519"

var errSignFormat = errors.New("errSignFormat")

type ed25519sign struct {
}

func newEd25519() Signer {
	return &ed25519sign{}
}

func (s *ed25519sign) GetLen() int {
	return 96
}

func (s *ed25519sign) GetPubKey(sign []byte) ([]byte, error) {
	if len(sign) != s.GetLen() {
		return nil, errSignFormat
	}
	return sign[0:32], nil
}

func (s *ed25519sign) Verify(data []byte, sign []byte) bool {
	_pubkey, err := s.GetPubKey(sign)
	if err != nil {
		return false
	}
	_sign := sign[32:]
	pubkey := new([32]byte)
	copy(pubkey[:], _pubkey)

	newsign := new([64]byte)
	copy(newsign[:], _sign)
	return ed25519.Verify(pubkey, data, newsign)
}
