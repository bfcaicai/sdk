package coolpad

import (
	"encoding/base64"
	"encoding/json"
	"github.com/astaxie/beego/httplib"
	"log"
	"math/big"
	"sdk/utils"
	"strings"
	"time"
)

const (
	appid            = "xxxxxx"
	appkey           = "xxxxxx"
	paykey           = "xxxxxx"
	access_token_url = "https://openapi.coolyun.com/oauth2/token"
	usr_info_url     = "https://openapi.coolyun.com/oauth2/api/get_user_info"
)

type AckCoolpadTokenInfo struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    string `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenId       string `json:"openid"`
}

type AckCoolpadUsrInfo struct {
	RtnCode     string `json:"rtn_code"`
	Sex         string `json:"sex"`
	NickName    string `json:"nickname"`
	Brithday    string `json:"brithday"`
	HighDefUrl  string `json:"highDefUrl"`
	HeadIconUrl string `json:"headIconUrl"`
}

type NotifyInfo struct {
	Exorderno string `json:"exorderno"`
	Transid   string `json:"transid"`
	Appid     string `json:"appid"`
	Waresid   int    `json:"waresid"`
	Feetype   int    `json:"feetype"`
	Money     int    `json:"money"`
	Count     int    `json:"count"`
	Result    int    `json:"result"`
	Transtype int    `json:"transtype"`
	Transtime string `json:"transtime"`
	Cpprivate string `json:"cpprivate"`
	Paytype   int    `json:"paytype"`
}

func coolpadAccessToken(code string) *AckCoolpadTokenInfo {
	var ack *AckCoolpadTokenInfo
	req := httplib.Get(access_token_url+
		"?grant_type=authorization_code&client_id="+
		appid+
		"&redirect_uri="+
		appkey+
		"&client_secret="+
		appkey+
		"&code="+code).SetTimeout(5*time.Second, 5*time.Second)
	if err := req.ToJSON(&ack); err != nil {
		log.Println(err)
		return nil
	}
	return ack
}

func CoolpadNotify(data, sign string) bool {
	log.Println("recv coolpad transaction order notify")
	md5str := utils.Md5String(data)
	decodeBaseKey := coolpadBase64KeyStr(paykey)
	decodeBaserep := strings.Replace(decodeBaseKey, "+", "#", -1)
	decodeBaseVec := strings.Split(decodeBaserep, "#")
	pkey := decodeBaseVec[0]
	mkey := decodeBaseVec[1]
	localsign := coolpadDecrypt(sign, pkey, mkey)
	if ok := strings.EqualFold(md5str, localsign); !ok {
		log.Println("sign error", "coolpad req sign:", md5str, "local sign:", localsign)
		return false
	}
	var (
		notify_inf NotifyInfo
	)
	if err := json.Unmarshal([]byte(data), &notify_inf); err != nil {
		log.Println(err)
		return false
	}
	cp_order := notify_inf.Exorderno
	cp_amt := notify_inf.Money
	log.Println(cp_order, cp_amt)
	//TODO your logic
	return true
}

func coolpadBase64KeyStr(pkey string) string {
	s1, _ := base64.StdEncoding.DecodeString(pkey)
	rs := []rune(string(s1))
	str := string(rs[40:])
	s2, _ := base64.StdEncoding.DecodeString(str)
	return string(s2)
}

func coolpadDecrypt(cryptograph, sd, sn string) string {
	d := string2big(sd, 10)
	n := string2big(sn, 10)
	arr := strings.Split(cryptograph, " ")
	length := len(arr)
	bigarr := make([]*big.Int, length)
	for i := 0; i < length; i++ {
		bigarr[i] = string2big(arr[i], 16)
	}
	arrbyte := coolpadDencodeRSA(bigarr, d, n)
	arrmsg := string2byte(arrbyte)
	return string(arrmsg)
}

func coolpadDencodeRSA(encodeM []*big.Int, d, n *big.Int) [][]byte {
	length := len(encodeM)
	dencodeM := make([][]byte, length)
	for i := 0; i < length; i++ {
		dencodeM[i] = encodeM[i].Exp(encodeM[i], d, n).Bytes()
	}
	return dencodeM
}

func string2byte(arr [][]byte) []byte {
	var b []byte
	for i := 0; i < len(arr); i++ {
		for j := 0; j < len(arr[i]); j++ {
			if arr[i][j] != ' ' {
				b = append(b, arr[i][j])
			}
		}
	}
	return b
}

func string2big(s string, radix int) *big.Int {
	ret := new(big.Int)
	ret.SetString(s, radix)
	return ret
}
