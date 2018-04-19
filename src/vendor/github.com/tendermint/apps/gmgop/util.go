package gmgop

import (
	"encoding/hex"
	"errors"
	"fmt"

	"dev.33.cn/33/crypto"
	_ "dev.33.cn/33/crypto/ed25519"
)

var ErrDupInstructionId = errors.New("ErrDupInstructionId")
var ErrSign = errors.New("error sign")
var ErrWrongMessageType = errors.New("wrong message type")
var ErrAssertExist = errors.New("assert exist")
var ErrAssertNotExist = errors.New("assert not exist")
var ErrSignWrongLength = errors.New("wrong length")
var ErrUidTooShort = errors.New("ErrUidTooShort")
var ErrUserExist = errors.New("user exist")
var ErrUserNotExist = errors.New("user not exist")
var ErrStorage = errors.New("storage error")
var ErrNotGop = errors.New("not gop")
var ErrNotBu = errors.New("not bu")
var ErrWrongPrice = errors.New("wrong price")
var ErrAssertNotEnough = errors.New("assert not enough")
var ErrAssertIsDelete = errors.New("assert is deleted")
var ErrAssertNotUsable = errors.New("assert is not usable")
var ErrContractExist = errors.New("contract exist")
var ErrContractNotExist = errors.New("contract not exist")
var ErrNoRight = errors.New("no right")
var ErrNotAdmin = errors.New("not admin")
var ErrWrongPubkey = errors.New("wrong pubkey")
var ErrNotApplicant = errors.New("not applicant")
var ErrCannotRecall = errors.New("can not recall")

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

func (app *GmgopApplication) saveReceipt(instructionId int64, receipt *Receipt) error {
	save, err := MarshalMessage(receipt)
	if err != nil {
		return err
	}
	app.state.Set(reqkey(instructionId), save)
	return err
}

func (app *GmgopApplication) checkInstructionId(instructionId int64) error {
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

func KeyAssert(id string) string {
	return "assert:" + fmt.Sprintf("%s", id)
}

func KeyContract(id string) string {
	return "contract:" + fmt.Sprintf("%s", id)
}

func KeyUser(id string) string {
	return "user:" + fmt.Sprintf("%s", id)
}

func isAdmin(str []byte) bool {
	_, ok := adminList[hex.EncodeToString(str)]
	if ok {
		return ok
	}
	return false
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
