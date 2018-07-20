package huawei

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"github.com/astaxie/beego/httplib"
	"log"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	gameprivatekey = []byte(`
-----BEGIN PRIVATE KEY-----
YOUR PRIVARE KEY
-----END PRIVATE KEY-----
`)
	paypublickey = []byte(`
-----BEGIN PUBLIC KEY-----
YOUR PAY PUBLIC KEY
-----END PUBLIC KEY-----
`)
)

const (
	appid            = "xxxxxx"
	appsecret        = "xxxxxx"
	payid            = "xxxxxx"
	cpid             = "xxxxxx"
	validPlayeridUrl = "https://gss-cn.game.hicloud.com/gameservice/api/gbClientApi"
)

type ValidPlayerid struct {
	GameAuthSign string
	PlayerLevel  string
	PlayerId     string
	Ts           int64
}

type AckValidPalyerSign struct {
	RtnCode int    `json:"rtnCode"`
	Ts      string `json:"ts"`
	RtnSign string `json:"rtnSign"`
}

func (data ValidPlayerid) ValidPlayeridSign() bool {
	dataMap := map[string]string{
		"method":      "external.hms.gs.checkPlayerSign",
		"appId":       appid,
		"cpId":        cpid,
		"ts":          strconv.Itoa(int(data.Ts)),
		"playerId":    data.PlayerId,
		"playerLevel": data.PlayerLevel,
		"playerSSign": data.GameAuthSign}
	var outdata *AckValidPalyerSign
	localsign, err := getsignature(dataMap, gameprivatekey)
	if err != nil {
		return false
	}
	dataMap["cpSign"] = localsign
	req := httplib.Post(validPlayeridUrl)
	req.SetTimeout(5*time.Second, 5*time.Second)
	for k, v := range dataMap {
		req.Param(k, v)
	}
	if err := req.ToJSON(&outdata); err != nil {
		log.Println(err)
		return false
	}
	if outdata.RtnCode != 0 {
		log.Println("check palyer sign eof out code:", outdata.RtnCode)
		return false
	}
	othMap := map[string]string{
		"rtnCode": strconv.Itoa(outdata.RtnCode),
		"ts":      outdata.Ts}
	localothsign, err := getsignature(othMap, gameprivatekey)
	if err != nil {
		log.Println(err)
		return false
	}
	if !strings.EqualFold(localothsign, outdata.RtnSign) {
		log.Println("check player sign out sign eof")
		log.Println(localothsign)
		log.Println(outdata.RtnSign)
		return false
	}
	return true
}

func getsignature(m map[string]string, key []byte) (string, error) {
	var strin []string
	strs := ""
	for k, _ := range m {
		switch k {
		case "cpSign":
			continue
		case "rtnSign":
			continue
		}
		strin = append(strin, k)
	}
	sort.Strings(strin)
	for k, v := range strin {
		v1, _ := m[v]
		s1 := v + "=" + url.QueryEscape(v1)
		if k == 0 {
			strs += s1
		} else {
			strs += "&" + s1
		}
	}
	log.Println(strs)
	block, _ := pem.Decode(key)
	if block == nil {
		log.Println("private key error")
		return "", errors.New("private key error")
	}
	priInterface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		log.Println(err)
		return "", err
	}
	h := sha256.New()
	h.Write([]byte(strs))
	buf := h.Sum(nil)
	sbyte, err := rsa.SignPKCS1v15(rand.Reader, priInterface.(*rsa.PrivateKey), crypto.SHA256, buf)
	if err != nil {
		log.Println(err)
		return "", err
	}
	return base64.StdEncoding.EncodeToString(sbyte), nil
}

func HuaweiNotify(dataMap map[string]string) int {
	outJsonCode := 1
	sign, _ := url.QueryUnescape(dataMap["sign"])
	extReserved, _ := url.QueryUnescape(dataMap["extReserved"])
	sysReserved, _ := url.QueryUnescape(dataMap["sysReserved"])
	dataMap["sign"] = sign
	dataMap["extReserved"] = extReserved
	dataMap["sysReserved"] = sysReserved
	if ok := checkpaysign(dataMap); !ok {
		log.Println("huawei notify rsa check error")
		return outJsonCode
	}
	if ok := strings.EqualFold("0", dataMap["result"]); !ok {
		log.Println("huawei pay status error")
		outJsonCode = 99
		return outJsonCode
	}
	cp_order, _ := strconv.Atoi(dataMap["requestId"])
	cp_amt, _ := strconv.ParseFloat(dataMap["amount"], 64)
	log.Println(cp_order, cp_amt)
	//TODO your logic
	outJsonCode = 0
	return outJsonCode
}

func checkpaysign(m map[string]string) bool {
	var strin []string
	sstr := ""
	for k, _ := range m {
		switch k {
		case "sign":
			continue
		case "signType":
			continue
		}
		strin = append(strin, k)
	}
	sort.Strings(strin)
	for k, v := range strin {
		v1, _ := m[v]
		if len(v1) > 0 {
			s1 := v + "=" + v1
			if k == 0 {
				sstr += s1
			} else {
				sstr += "&" + s1
			}
		}
	}
	log.Println(sstr)
	block, _ := pem.Decode(paypublickey)
	if block == nil {
		log.Println("public key error")
		return false
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		log.Println(err)
		return false
	}
	pub := pubInterface.(*rsa.PublicKey)
	h := sha256.New()
	h.Write([]byte(sstr))
	buf := h.Sum(nil)
	repSign := strings.Replace(m["sign"], " ", "+", -1)
	byte64, _ := base64.StdEncoding.DecodeString(repSign)
	err = rsa.VerifyPKCS1v15(pub, crypto.SHA256, buf, byte64)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}
