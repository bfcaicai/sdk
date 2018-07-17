package ysdk

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/httplib"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	ysdk_svr_url   = "https://ysdk.qq.com"
	app_paykey     = "xxxxxx"
	qq_appid       = "xxxxxx"
	qq_appkey      = "xxxxxx"
	wechat_appid   = "xxxxxx"
	wechat_appkey  = "xxxxxx"
	QQ_check_api   = "/auth/qq_check_token"
	WX_check_api   = "/auth/wx_check_token"
	GetBalance_api = "/mpay/get_balance_m"
	BuyItem_api    = "/mpay/pay_m"
	BuyCancel_api  = "/mpay/cancel_pay_m"
	keng_str       = "/v3/r"
)

type MsgYSDKUsrToken struct {
	MerId       string
	Openid      string
	AccessToken string
	Type        int
	DeviceCode  string
	Sign        string
}

type OutData struct {
	Ret int    `json:"ret"`
	Msg string `json:"msg"`
}

type YSDKtssinfo struct {
	Innerproductid         string `json:"innerproductid"`
	Begintime              string `json:"begintime"`
	Endtime                string `json:"endtime"`
	Paychan                string `json:"paychan"`
	Paysubchan             int    `json:"paysubchan"`
	Autopaychan            string `json:"autopaychan"`
	Autopaysubchan         int    `json:"autopaysubchan"`
	Grandtotal_opendays    int    `json:"grandtotal_opendays"`
	Grandtotal_presentdays int    `json:"grandtotal_presentdays"`
	First_buy_time         string `json:"first_buy_time"`
	Extend                 string `json:"extend"`
}

type YSDKGetBalanceSuccess struct {
	Ret         int            `json:"ret"`
	Balance     int            `json:"balance"`
	Gen_balance int            `json:"gen_balance"`
	First_save  int            `json:"first_save"`
	Save_amt    int            `json:"save_amt"`
	Gen_expire  int            `json:"gen_expire"`
	Tss_list    []*YSDKtssinfo `json:"tss_list"`
}

type MsgGetBalance struct {
	MerId    string
	Openid   string
	Openkey  string
	Appid    string
	Ts       string
	Sig      string
	Pf       string
	Pfkey    string
	Zoneid   string
	Userip   string
	Format   string
	PayToken string
	Type     int
	Sign     string
}

type ItemBuyInfo struct {
	MerId    string
	Openid   string
	Openkey  string
	Appid    string
	Ts       string
	Sig      string
	Pf       string
	Pfkey    string
	Zoneid   string
	Amt      int
	Billno   string
	Userip   string
	Format   string
	PayToken string
	Type     int
	Sign     string
}

type ItemBuyOk struct {
	Ret    int    `json:"ret"`
	Billno string `json:"billno"`
}

func (t *MsgGetBalance) GetBalance() (int, int) {
	_map := map[string]string{
		"openid":    t.Openid,
		"openkey":   t.Openkey,
		"pay_token": t.PayToken,
		"appid":     t.Appid,
		"ts":        t.Ts,
		"pf":        t.Pf,
		"pfkey":     t.Pfkey,
		"zoneid":    t.Zoneid}
	querystr := sortstr(_map)
	yuanstr := "GET&" + url.QueryEscape(keng_str+GetBalance_api) + "&" + url.QueryEscape(querystr)
	log.Println(yuanstr)
	keystr := app_paykey + "&"
	log.Println(keystr)
	_map["sig"] = url.QueryEscape(sigstr(keystr, yuanstr))
	log.Println(_map["sig"])
	queryurl, _ := url.Parse(ysdk_svr_url + GetBalance_api)
	cookieJar, _ := cookiejar.New(nil)
	cookieJar.SetCookies(queryurl, setcookies(t.Type, GetBalance_api))
	client := &http.Client{Jar: cookieJar}
	resp, err := client.Get(ysdk_svr_url +
		GetBalance_api +
		"?" + querystr +
		"&sig=" + _map["sig"])
	if err != nil {
		log.Println(err)
		return 0, 0
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return 0, 0
	}
	str := string(body)
	log.Println(str)
	if strings.Contains(str, "msg") {
		log.Println(str)
		return 0, 0
	}
	var out *YSDKGetBalanceSuccess
	if err := json.Unmarshal(body, &out); err != nil {
		log.Println(err)
		return 0, 0
	}
	return out.Balance, out.Save_amt
}

func (t *ItemBuyInfo) ItemBuy() bool {
	_map := map[string]string{
		"openid":    t.Openid,
		"openkey":   t.Openkey,
		"pay_token": t.PayToken,
		"appid":     t.Appid,
		"ts":        t.Ts,
		"pf":        t.Pf,
		"pfkey":     t.Pfkey,
		"zoneid":    t.Zoneid,
		"amt":       strconv.Itoa(t.Amt),
		"billno":    t.Billno}
	querystr := sortstr(_map)
	yuanstr := "GET&" + url.QueryEscape(keng_str+BuyItem_api) + "&" + url.QueryEscape(querystr)
	log.Println(yuanstr)
	keystr := app_paykey + "&"
	log.Println(keystr)
	_map["sig"] = url.QueryEscape(sigstr(keystr, yuanstr))
	log.Println(_map["sig"])
	queryurl, _ := url.Parse(ysdk_svr_url + BuyItem_api)
	cookieJar, _ := cookiejar.New(nil)
	cookieJar.SetCookies(queryurl, setcookies(t.Type, BuyItem_api))
	client := &http.Client{Jar: cookieJar}
	reqfullurl := ysdk_svr_url +
		BuyItem_api +
		"?" + querystr +
		"&sig=" + _map["sig"]
	log.Println(reqfullurl)
	resp, err := client.Get(reqfullurl)
	if err != nil {
		log.Println(err)
		return false
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return false
	}
	str := string(body)
	log.Println(str)
	if strings.Contains(str, "msg") {
		log.Println(str)
		return false
	}
	var out *ItemBuyOk
	if err := json.Unmarshal([]byte(str), &out); err != nil {
		log.Println(err)
		return false
	}
	if !strings.EqualFold(t.Billno, out.Billno) {
		log.Println("billno not equal")
		return false
	}
	return true
}

func (t *ItemBuyInfo) BuyCancel() {
	_map := map[string]string{
		"openid":    t.Openid,
		"openkey":   t.Openkey,
		"pay_token": t.PayToken,
		"appid":     t.Appid,
		"ts":        t.Ts,
		"pf":        t.Pf,
		"pfkey":     t.Pfkey,
		"zoneid":    t.Zoneid,
		"amt":       strconv.Itoa(t.Amt),
		"billno":    t.Billno}
	querystr := sortstr(_map)
	yuanstr := "GET&" + url.QueryEscape("/v3/r"+BuyCancel_api) + "&" + url.QueryEscape(querystr)
	log.Println(yuanstr)
	keystr := app_paykey + "&"
	log.Println(keystr)
	_map["sig"] = url.QueryEscape(sigstr(keystr, yuanstr))
	log.Println(_map["sig"])
	queryurl, _ := url.Parse(ysdk_svr_url + BuyCancel_api)
	cookieJar, _ := cookiejar.New(nil)
	cookieJar.SetCookies(queryurl, setcookies(t.Type, BuyCancel_api))
	client := &http.Client{Jar: cookieJar}
	resp, err := client.Get(ysdk_svr_url +
		BuyCancel_api +
		"?" + querystr +
		"&sig=" + _map["sig"])
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}
	str := string(body)
	log.Println(str)
	if strings.Contains(str, "msg") {
		log.Println(str)
		return
	}
}

func ValidateAccessToken(ntype int, openid, token, usrip string) bool {
	var outdata *OutData
	appid, appkey, api := qq_appid, qq_appkey, QQ_check_api
	if ntype == 1 {
		appid, appkey, api = wechat_appid, wechat_appkey, WX_check_api
	}
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	sig := md5String(appkey + timestamp)
	log.Println("query sig:", sig)
	url_param := "?timestamp=" + timestamp + "&appid=" + appid +
		"&sig=" + sig + "&openid=" + openid + "&openkey=" + token + "&userip=" + usrip
	req := httplib.Get(ysdk_svr_url+api+url_param).SetTimeout(5*time.Second, 5*time.Second)
	if err := req.ToJSON(&outdata); err != nil {
		log.Println(err)
		return false
	}
	if outdata.Ret != 0 {
		log.Println(outdata.Ret, outdata.Msg)
		return false
	}
	return true
}

func sortstr(m map[string]string) string {
	var strin []string
	keys := ""
	for k, _ := range m {
		switch k {
		case "sig":
			continue
		}
		strin = append(strin, k)
	}
	sort.Strings(strin)
	for k, v := range strin {
		v1, _ := m[v]
		s1 := v + "=" + v1
		if k == 0 {
			keys += s1
		} else {
			keys += "&" + s1
		}
	}
	log.Println(keys)
	return keys
}

func sigstr(key, str string) string {
	k := []byte(key)
	mac := hmac.New(sha1.New, k)
	mac.Write([]byte(str))
	arrbyte := mac.Sum(nil)
	sig_str := base64.StdEncoding.EncodeToString(arrbyte)
	log.Println(sig_str)
	return sig_str
}

func setcookies(ntype int, apistr string) []*http.Cookie {
	var (
		cookiels []*http.Cookie
		cookie_sessionid,
		cookie_sessiontype,
		cookie_orgloc *http.Cookie
	)
	if ntype == 0 {
		cookie_sessionid = &http.Cookie{
			Name:  "session_id",
			Value: "openid"}
		cookiels = append(cookiels, cookie_sessionid)
		cookie_sessiontype = &http.Cookie{
			Name:  "session_type",
			Value: "kp_actoken"}
		cookiels = append(cookiels, cookie_sessiontype)
	} else {
		cookie_sessionid = &http.Cookie{
			Name:  "session_id",
			Value: "hy_gameid"}
		cookiels = append(cookiels, cookie_sessionid)
		cookie_sessiontype = &http.Cookie{
			Name:  "session_type",
			Value: "wc_actoken"}
		cookiels = append(cookiels, cookie_sessiontype)
	}
	cookie_orgloc = &http.Cookie{
		Name:  "org_loc",
		Value: url.QueryEscape(keng_str + apistr)}
	cookiels = append(cookiels, cookie_orgloc)
	return cookiels
}

func md5String(md5Str string) string {
	md5 := md5.New()
	io.WriteString(md5, md5Str)
	return fmt.Sprintf("%x", md5.Sum(nil))
}
