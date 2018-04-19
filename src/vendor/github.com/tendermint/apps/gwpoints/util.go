package gwpoints

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"

	"dev.33.cn/33/crypto"
	_ "dev.33.cn/33/crypto/ed25519"
)

var ErrDupInstructionId = errors.New("ErrDupInstructionId")
var ErrSign = errors.New("error sign")
var ErrWrongMessageType = errors.New("wrong message type")
var ErrSignWrongLength = errors.New("wrong length")
var ErrUidTooShort = errors.New("ErrUidTooShort")
var ErrUserExist = errors.New("user exist")
var ErrUserNotExist = errors.New("user not exist")
var ErrStorage = errors.New("storage error")
var ErrNoRight = errors.New("no right")
var ErrNotAdmin = errors.New("not admin")
var ErrWrongPubkey = errors.New("wrong pubkey")
var ErrPlatformExist = errors.New("platform exist")
var ErrPlatformNotExist = errors.New("platform not exist")
var ErrWrongPoints = errors.New("wrong points value")
var ErrPointsNotEnough = errors.New("points not enough")
var ErrWrongCompanyId = errors.New("wrong company id")

func pad32(i int32) string {
	return fmt.Sprintf("%010d", i)
}

func pad64(i int64) string {
	return fmt.Sprintf("%020d", i)
}

var adminList = map[string]bool{
	"b15a4f6c5c1163b5f80715c9bd87d5118ec4b5668cb29f148eeceec61ddeadc2": true,
}

var bankList = map[string]bool{
	"4a246cd2a3f41b2bc1d071d2db159a388cd6f5c3547ea592d401f270073133d7": true,
}

func isBank(uid []byte) bool {
	_, ok := bankList[hex.EncodeToString(uid)]
	return ok
}

func (app *GwpointsApplication) saveReceipt(instructionId int64, receipt *Receipt) error {
	save, err := MarshalMessage(receipt)
	if err != nil {
		return err
	}
	app.state.Set(reqkey(instructionId), save)
	return err
}

func (app *GwpointsApplication) checkInstructionId(instructionId int64) error {
	key := reqkey(instructionId)
	_, _, exists := app.state.Get(key)
	if !exists {
		return nil
	}
	return ErrDupInstructionId
}

func reqkey(n int64) []byte {
	s := fmt.Sprintf("reqkey_%d", n)
	return []byte(s)
}

func KeyUser(id int64) string {
	return "user:" + fmt.Sprintf("%d", id)
}

func KeyPlatform() string {
	return "platform"
}

func isOriginalAdmin(str []byte) bool {
	_, ok := adminList[hex.EncodeToString(str)]
	if ok {
		return ok
	}
	return false
}

func (app *GwpointsApplication) isAdmin(pubkey []byte) bool {
	_, value, exists := app.state.Get([]byte(KeyPlatform()))
	if !exists {
		return false
	}
	var platform Platform
	err := UnmarshalMessage(value, &platform)
	if err != nil {
		return false
	}
	if !bytes.Equal(platform.Pubkey, pubkey) {
		return false
	}
	return true
}

func CheckSign(data []byte, uid []byte, sign []byte) error {
	c, err := crypto.New("ed25519")
	if err != nil {
		return err
	}
	if len(sign) != 64 {
		return ErrSignWrongLength
	}
	sig, err := c.SignatureFromBytes(sign)
	if err != nil {
		return err
	}
	pub, err := c.PubKeyFromBytes(uid[:])
	if err != nil {
		return err
	}
	if !pub.VerifyBytes(data, sig) {
		return ErrSign
	}
	return nil
}

func Signdata(privKey []byte, data []byte) []byte {
	c, err := crypto.New("ed25519")
	if err != nil {
		panic(err)
	}
	priv, err := c.PrivKeyFromBytes(privKey)
	if err != nil {
		panic(err)
	}
	sig := priv.Sign(data)
	return sig.Bytes()
}
