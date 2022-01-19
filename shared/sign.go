package utils

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"net/url"
	"sort"
)

func ParaFilter(args map[string]string) map[string]string {
	// 除去数组中的空值和签名参数
	var newArgs = map[string]string{}
	if args == nil || len(args) <= 0 {
		return newArgs
	}
	for k, v := range args {
		if v == "" || k == "signature" || k == "signMethod" {
			continue
		} else {
			newArgs[k] = v
		}

	}
	return newArgs
}

func CreateLinkString(args map[string]string, bsort bool, encode bool) string {
	var keys = []string{}
	for k := range args {
		keys = append(keys, k)
	}
	if bsort {
		sort.Strings(keys)
	}
	var prestr string
	for i := 0; i < len(keys); i++ {
		k := keys[i]
		v := args[k]
		if encode {
			v = url.QueryEscape(v)
		}
		if i == len(keys)-1 { //拼接时，不包括最后一个&字符
			prestr = prestr + k + "=" + v
		} else {
			prestr = prestr + k + "=" + v + "&"
		}
	}
	return prestr
}

func GETHMACSHA1(keyStr, value string) string {

	key := []byte(keyStr)
	mac := hmac.New(sha1.New, key)
	mac.Write([]byte(value))
	//进行base64编码
	res := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	return res
}
