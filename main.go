package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net"
	"fmt"
	"time"
	"net/http"
	"strings"
	"regexp"

	"github.com/elazarl/goproxy"
	"github.com/pkg/errors"
	log_std "github.com/xxzl0130/AntiRivercrab/pkg/log"
	cipher "github.com/xxzl0130/GF_cipher"
	"github.com/xxzl0130/AntiRivercrab/pkg/util"
)

type AntiRivercrab struct {
	log  log_std.Logger
	sign string
	pattern string
	replacement string
}

func main() {
	log, err := log_std.New(fmt.Sprintf("AntiRivercrab.%d.log", time.Now().Unix()))
	if err != nil {
		panic(fmt.Sprintf("%+v", err))
	}

	ar := &AntiRivercrab{
		log:  log,
		sign : "",
		pattern : "\"naive_build_gun_formula\":\"(\\d+:\\d+:\\d+:\\d+)?\"",
		replacement : "\"naive_build_gun_formula\":\"33:33:33:33\"",
	}
	if err := ar.Run(); err != nil {
		ar.log.Fatalf("程序启动失败 -> %+v", err)
	}
}

func (ar *AntiRivercrab) Run() error {
	localhost, err := ar.getLocalhost()
	if err != nil {
		ar.log.Fatalf("获取代理地址失败 -> %+v", err)
	}

	ar.log.Tipsf("代理地址 -> %s:%d", localhost, 8888)

	srv := goproxy.NewProxyHttpServer()
	srv.Logger = new(util.NilLogger)
	srv.OnResponse(ar.condition()).DoFunc(ar.onResponse)

	if err := http.ListenAndServe(":8888", srv); err != nil {
		ar.log.Fatalf("启动代理服务器失败 -> %+v", err)
	}

	return nil
}

func (ar *AntiRivercrab) onResponse(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
	type Uid struct {
		Sign            string `json:"sign"`
	}

	ar.log.Infof("处理请求响应 -> %s", path(ctx.Req))

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ar.log.Errorf("读取响应数据失败 -> %+v", err)
		resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		return resp
	}
	if strings.HasSuffix(ctx.Req.URL.Path,"/Index/getDigitalSkyNbUid"){
		data, err := cipher.AuthCodeDecodeB64Default(string(body)[1:])
		if err != nil {
			ar.log.Errorf("解析Uid数据失败 -> %+v", err)
			resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))
			return resp
		}
		uid := Uid{}
		if err := json.Unmarshal([]byte(data), &uid); err != nil {
			ar.log.Errorf("解析JSON数据失败 -> %+v", err)
			resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))
			return resp
		}
		ar.sign = uid.Sign
		ar.log.Infof("解析Uid成功")
		resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		return resp
	} else if strings.HasSuffix(ctx.Req.URL.Path,"/Index/index"){
		data, err := cipher.AuthCodeDecodeB64(string(body)[1:], ar.sign, true)
		if err != nil {
			ar.log.Errorf("解析用户数据失败 -> %+v", err)
			resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))
			return resp
		}
		ar.log.Infof("解析用户数据成功")
		body = []byte(regexp.MustCompile(ar.pattern).ReplaceAll([]byte(data), []byte(ar.replacement)))
		if(err != nil){
			ar.log.Errorf("打包用户数据失败 -> %+v", err)
			resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))
			return resp
		}
		ar.log.Tipsf("AntiRivercrab行动成功")
	}
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	return resp
}

func (ar *AntiRivercrab) condition() goproxy.ReqConditionFunc {
	return func(req *http.Request, ctx *goproxy.ProxyCtx) bool {
		ar.log.Infof("请求 -> %s", path(req))
		if strings.HasSuffix(req.Host, "ppgame.com") {
			if strings.HasSuffix(req.URL.Path, "/Index/index") || strings.HasSuffix(req.URL.Path, "/Index/getDigitalSkyNbUid"){
				ar.log.Infof("请求通过 -> %s", path(req))
				return true
			}
		}
		return false
	}
}

func (ar *AntiRivercrab) getLocalhost() (string, error) {
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
