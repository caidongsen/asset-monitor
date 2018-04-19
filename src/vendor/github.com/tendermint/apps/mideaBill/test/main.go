package main

import (
	//"encoding/base64"
	"encoding/hex"
	"fmt"
)

func main() {
	//s := "erEBChZNNTkzMDg5NzgyMDE3MTIwODAwMDNQElIKFk02NzA1MTg2MDIwMTcxMTI3MDAwNVMQgNrECRon6auY5bmz5biC57qi5peX5ZWG5Zy65pyJ6ZmQ6LSj5Lu75YWs5Y+4IgoxMTEzNTI0OC1YEkMKFk02NzA1MTg2MDIwMTcxMTI3MDAwNlMQgK3iBBoY5a6J5rO95Y6/576O5rqQ55S15Zmo5Z+OIgpNQTBIMjNQMS050gEJZmgwMDAwMDAx2AGFk9yDkJK18/8B4gEgqwP/vxyEDNg5pTgOWitYZzeI+inzm9zgHWQjVdevte7qAUB2FnkJmDCyYECljdupxpKCJWwgrRx/zWuHtAM6XGU3AB+XdLGeHaakw//y8JmCLdP0c3uubt8kg0nYcPS7oCwA8AEP"
//s:= "EhcKCjU5OTkzMTg0LTcSCeadjuS6muWugboBBkFsdHVtbsABlK7G/8/5wxnKASA3kGctLuYH9LMUWhdOoI3OnmIv1X9Y5rfNAnUhKejpXtIBQEqYBepgTbXUJLepy17MKXPR/3e08/VYQkISylFbbWeMeGRqM/gcdy8L8rxkMAXnrKm3B6DZiymqK50l2C0F4wjYAQI="
	//buf, err := base64.StdEncoding.DecodeString(s)

	s := "a201440a164d313238333930333232303137313231353030303153121ee7be8ee79a84e59586e4b89ae4bf9de79086e69c89e99990e585ace58fb81a0a33353939323034312d58d201086c696a7a30303831d801a89b9c8080f2a3c5ff01e20120ad7f9d0578a323d4225d39d56a543132e37a8c2a8b5c3f9c69a59ec87fe0297fea014078541c250718728da2e3d866426b193d851aa228183953c17407aa77f67b7e5a383f776dbdb67a38b83c01378142325356230750cf952fec4ea1b7761cc89e01f00114f80181b5ced105"

	buf, err := hex.DecodeString(s)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(buf)
	
	var req Request
        err = UnmarshalMessage(buf, &req)
        if err != nil {
		fmt.Println(err)
                return 
        }
fmt.Println(req)
	err = req.CheckSign()
        if err != nil {
                fmt.Println(err)
        }


	sub := req.GetBillTotalFinancing()
	fmt.Println(sub)
	

}
