package smsc

import (
	"encoding/hex"
	"testing"
)

var gwpointsApp *GwpointsApplication
var adminPrivkey *[64]byte
var userKey *[64]byte
var bankKey *[64]byte

func init() {
	gwpointsApp = NewGwpointsApplication()
	//adminPrivkey = HexToPrivkey("90b289fda1fb0439158f837bbe60cc1ec99616dd0bc6335d6fd0bf3d22888e20b15a4f6c5c1163b5f80715c9bd87d5118ec4b5668cb29f148eeceec61ddeadc2")
	//bankKey = HexToPrivkey("86c8b35d96a2968a9b82e67ac19e1312c0618ffd3e93f75510c7f10e192b75fb4a246cd2a3f41b2bc1d071d2db159a388cd6f5c3547ea592d401f270073133d7")
	//userKey = HexToPrivkey("8a1a5a9ce333e10704e58bc331f9693ee2e126a98dc365bed57020bedc22a2cbf6e008a45ffc21a902ed34b4b4de04b743e20677855cc278c70192fb1f476f34")
}

func TestCaseUserCreate(t *testing.T) {
	req := &RequestUserCreate{}
	req.UserUid = 111111
	req.UserPubkey, _ = hex.DecodeString("f6e008a45ffc21a902ed34b4b4de04b743e20677855cc278c70192fb1f476f34")
	req.Info, _ = hex.DecodeString("xxx")
	request := &Request{}
	request.Value = &Request_UserCreate{req}
	request.Uid = 0
	request.InstructionId = 111
	request.Pubkey, _ = hex.DecodeString("f6e008a45ffc21a902ed34b4b4de04b743e20677855cc278c70192fb1f476f34")
	request.ActionId = MessageType_MsgUserCreate
	data2, err := MarshalMessage(request)
	if err != nil {
		t.Fatal(err)
	}
	priv, _ := hex.DecodeString("8a1a5a9ce333e10704e58bc331f9693ee2e126a98dc365bed57020bedc22a2cbf6e008a45ffc21a902ed34b4b4de04b743e20677855cc278c70192fb1f476f34")
	request.Sign = Signdata(priv, data2)

	data2, err = MarshalMessage(request)
	if err != nil {
		t.Fatal(err)
	}
	err = gwpointsApp.Check(data2)
	if err != nil {
		t.Fatal(err)
	}
	res, err := gwpointsApp.Exec(data2)
	if err != nil {
		t.Fatal(err)
	}
	var resp Response
	err = UnmarshalMessage(res, &resp)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(resp)
}
