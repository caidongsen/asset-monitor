package mideaBill

import (
	//"bytes"
	"encoding/hex"
	"errors"
	"fmt"

	"dev.33.cn/33/crypto"
	_ "dev.33.cn/33/crypto/ed25519"
	"github.com/tendermint/ed25519"
)

var ErrPlatformIsInit = errors.New("platform is init")
var ErrNotOriginalAdmin = errors.New("not original admin")
var ErrPubkeyNotMatch = errors.New("pubkey not match")
var ErrGetValueIsNull = errors.New("get value is null")
var ErrUserNameIsNull = errors.New("username is null")
var ErrContactNameIsNull = errors.New("contactname is null")
var ErrOperatorIsNull = errors.New("operator is null")
var ErrUserNameIsNotSDKUser = errors.New("username is not sdkuser")
var ErrNotAdmin = errors.New("not admin")
var ErrUserPublicKey = errors.New("userpublickey is err")
var ErrUserExist = errors.New("user exist")
var ErrUserNotExist = errors.New("user not exist")
var ErrSDKUserNameIsNull = errors.New("sdk username is null")
var ErrSDKUserNameIsNotMatch = errors.New("sdk username is not match")
var ErrSDKPublicKey = errors.New("sdk publickey is err")

var ErrEntExist = errors.New("ent exist")
var ErrEntNotExist = errors.New("ent not exist")
var ErrEntNameIsNull = errors.New("entName is null")
var ErrEntCodeIsNull = errors.New("entCode is null")
var ErrEntTypeNotMatch = errors.New("entType not match")
var ErrMideaDraftIdIsNull = errors.New("mideaDraftId is null")
var ErrMideaDraftAmountIsNull = errors.New("mideaDraftAmount is null")
var ErrIssueBillDayIsErr = errors.New("issueBillDay is err")
var ErrExpireDayIsErr = errors.New("expireDay is err")
var ErrPayNumIsNull = errors.New("payNum is null")
var ErrRecvBillEntNameIsNull = errors.New("recvBillEntName is null")
var ErrRecvBillEntCodeIsNull = errors.New("recvBillEntCode is null")
var ErrEntNameNotMatch = errors.New("entName not match")
var ErrEntCodeNotMatch = errors.New("entCode not match")
var ErrWaitRecvBillEntNameIsNull = errors.New("waitRecvBillEntName is null")
var ErrWaitRecvBillEntCodeIsNull = errors.New("waitRecvBillEntCode is null")
var ErrWaitRecvBillEntCodeNotMatch = errors.New("waitRecvBillEntCode not match")

var ErrBillIsExists = errors.New("bill is exists")
var ErrBillIsNotExists = errors.New("bill is not exists")
var ErrRecvBillPublicKeyNotMatch = errors.New("recvBillPublicKey not match")
var ErrIssueBillPublicKeyNotMatch = errors.New("issueBillPublicKey not match")
var ErrBillStateNotMatch = errors.New("billState not match")
var ErrBillSubListIsNull = errors.New("billSubList is null")
var ErrMideaDraftAmountNotMatch = errors.New("mideaDraftAmount not match")
var ErrWaitRecvBillPublicKeyNotMatch = errors.New("waitRecvBillPublicKey not match")
var ErrExpireDayNotMatch = errors.New("expireDay not match")


var ErrDupInstructionId = errors.New("ErrDupInstructionId")
var ErrSign = errors.New("error sign")
var ErrWrongMessageType = errors.New("wrong message type")
var ErrSignWrongLength = errors.New("wrong length")
var ErrStorage = errors.New("storage error")


// Java-SDK用户名
var SdkUser string = "sdkuser"

var adminList = map[string]bool{
	"b15a4f6c5c1163b5f80715c9bd87d5118ec4b5668cb29f148eeceec61ddeadc2": true,
}

func isOriginalAdmin(str []byte) bool {
	_, ok := adminList[hex.EncodeToString(str)]
	return ok
}

func KeyReq(n int64) []byte {
	s := fmt.Sprintf("keyreq_%d", n)
	return []byte(s)
}

func KeyPlatform() []byte {
	return []byte("platform")
}

func KeyUser(userName string) []byte {
	return []byte("user:" + userName)
}

func KeyEnt(entName string) []byte {
	return []byte("ent:" + entName)
}

func KeyDraft(draftId string) []byte {
	return []byte("draft:" + draftId)
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

func GetEntPublicKey(entCode string) []byte {
	return GetPublicKey(GetPrivateKey(entCode))
}

func GetPrivateKey(entCode string) []byte {
	h := New256()
	h.Write([]byte(entCode))
	keccakhash := h.Sum(nil)
	//ss := hex.EncodeToString(keccakhash)
	//fmt.Println(ss)
	return keccakhash
}

func GetPublicKey(priKey []byte) []byte {
	privKeyBytes := new([64]byte)
	copy(privKeyBytes[:32], priKey[:32])
	pubbuf := ed25519.MakePublicKey(privKeyBytes)
	public := pubbuf[:]
	return public
}