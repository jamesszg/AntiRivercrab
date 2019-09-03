package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"strings"

	"github.com/elazarl/goproxy"
	log "github.com/xxzl0130/AntiRivercrab/pkg/log"
	"github.com/xxzl0130/AntiRivercrab/pkg/cipher"
	"github.com/xxzl0130/AntiRivercrab/pkg/util"
)

var sign string

func main() {
	localhost, err := getLocalhost()
	if err != nil {
		log.Fatalf("获取代理地址失败 -> %+v", err)
	}

	log.Tipsf("代理地址 -> %s:%d", localhost, 8888)

	srv := goproxy.NewProxyHttpServer()
	srv.Logger = new(util.NilLogger)
	srv.OnResponse(condition()).DoFunc(onResponse)

	if err := http.ListenAndServe(":8888", srv); err != nil {
		log.Fatalf("启动代理服务器失败 -> %+v", err)
	}
}

func onResponse(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
	type Uid struct {
		Sign            string `json:"sign"`
	}

	log.Infof("处理请求响应 -> %s", path(ctx.Req))

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("读取响应数据失败 -> %+v", err)
		return resp
	}
	if strings.HasSuffix(ctx.Req.URL.Path,"/Index/getDigitalSkyNbUid"){
		data, err := cipher.AuthCodeDecodeB64(string(body)[1:], "yundoudou", true)
		if err != nil {
			log.Errorf("解析Uid数据失败 -> %+v", err)
			return resp
		}
		uid := Uid{}
		if err := json.Unmarshal([]byte(data), &uid); err != nil {
			log.Errorf("解析JSON数据失败 -> %+v", err)
			return resp
		}
		sign = uid.Sign
		return resp
	} else if strings.HasSuffix(ctx.Req.URL.Path,"/Index/index"){
		data, err := cipher.AuthCodeDecodeB64(string(body)[1:], sign, true)
		if err != nil {
			log.Errorf("解析用户数据失败 -> %+v", err)
			return resp
		}
		body = []byte(strings.Replace(data,"\"naive_build_gun_formula\": \"\"","\"naive_build_gun_formula\":\"33:33:33:33\"",1))
	}
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	return resp
}

func condition() goproxy.ReqConditionFunc {
	return func(req *http.Request, ctx *goproxy.ProxyCtx) bool {
		log.Infof("请求 -> %s", path(req))
		if strings.HasSuffix(req.Host, "ppgame.com") || strings.HasSuffix(req.Host, "sn-game.txwy.tw") {
			if strings.HasSuffix(req.URL.Path, "/Index/index") || strings.HasSuffix(req.URL.Path, "/Index/getDigitalSkyNbUid"){
				log.Infof("请求通过 -> %s", path(req))
				return true
			}
		}
		return false
	}
}

func getLocalhost() (string, error) {
	conn, err := net.Dial("tcp", "www.baidu.com:80")
	if err != nil {
		return "", errors.WithMessage(err, "连接 www.baidu.com:80 失败")
	}
	host, _, err := net.SplitHostPort(conn.LocalAddr().String())
	if err != nil {
		return "", errors.WithMessage(err, "解析本地主机地址失败")
	}
	return host, nil
}

func path(req *http.Request) string {
	if req.URL.Path == "/" {
		return req.Host
	}
	return req.Host + req.URL.Path
}
