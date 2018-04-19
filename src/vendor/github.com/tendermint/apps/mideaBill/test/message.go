package main

import (
"fmt"
"encoding/hex"
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
fmt.Println("sign")
fmt.Println(hex.EncodeToString(sign))
	req.Sign = nil
	data, err := MarshalMessage(req)
	if err != nil {
		return err
	}
fmt.Println(hex.EncodeToString(data))
fmt.Println("sign")
	return CheckSign(data, req.Pubkey, sign)
}

func (req *Request) Sign11() error {
	sign := req.GetSign()
	req.Sign = nil
fmt.Println("sign bre---------------")
fmt.Println(sign)
fmt.Println(req)
fmt.Println("sign bre---------------")
	data, err := MarshalMessage(req)
	if err != nil {
		return err
	}
	prv, err := hex.DecodeString("087947d820a476a3d243bd13a193fc96af4a02a4fbe024a19a89fb7ac3b68af0")
	if err != nil {
		return err
	}
	var p [64]byte
	copy(p[:32],prv)
	copy(p[32:],req.Pubkey)
	si := Signdata(p[:], data)
	fmt.Println("sign-----------")
	fmt.Println(si)
	fmt.Println(sign)
	fmt.Println("sign-----------")
	pub := GetPublicKey(p[:])
	fmt.Println(hex.EncodeToString(pub))
	return nil
}
