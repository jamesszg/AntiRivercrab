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
	"path/filepath"

	"github.com/elazarl/goproxy"
	"github.com/pkg/errors"
	log_std "github.com/xxzl0130/AntiRivercrab/pkg/log"
	cipher "github.com/xxzl0130/AntiRivercrab/GF_cipher"
	"github.com/xxzl0130/AntiRivercrab/pkg/util"
)

type SignInfo struct {
	sign string
	time int64
}

type AntiRivercrab struct {
	log  log_std.Logger
	sign map[string]SignInfo
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
		sign : make(map[string]SignInfo) ,
		pattern : "\"naive_build_gun_formula\":\"(\\d+:\\d+:\\d+:\\d+)?\"",
		replacement : "\"naive_build_gun_formula\":\"33:33:33:33\"",
	}
	if err := ar.Run(); err != nil {
		ar.log.Fatalf("程序启动失败 -> %+v", err)
	}
}

func (ar *AntiRivercrab) watchdog(){
	for{
		time.Sleep(time.Second * 60 * 10)
		now := time.Now().Unix()
		for k,v := range ar.sign{
			if now - v.time > (60 * 10){
				fmt.Printf("delete %s\n",k)
				delete(ar.sign,k)
			}
		}
	}
}

func (ar *AntiRivercrab) Run() error {
	localhost, err := ar.getLocalhost()
	if err != nil {
		ar.log.Fatalf("获取代理地址失败 -> %+v", err)
		return err
	}

	ar.log.Tipsf("代理地址 -> %s:%d", localhost, 8888)

	proxy := goproxy.NewProxyHttpServer()
	proxy.Logger = new(util.NilLogger)
	proxy.OnResponse(ar.condition()).DoFunc(ar.onResponse)

	proxy.OnRequest(ar.block()).DoFunc(
		func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			return r, goproxy.NewResponse(r, goproxy.ContentTypeText, http.StatusForbidden, "abuse of this service is not allowed!")
		})

	srv := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		Addr:         ":8888",
		Handler:      proxy,
	}

	go srv.ListenAndServe()

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("."+string(filepath.Separator)+"PACFile")))
	fileSrv := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		Addr:         ":3000",
		Handler:      mux,
	}
	go ar.watchdog()
	fileSrv.ListenAndServe()
	return nil
}

func (ar *AntiRivercrab) onResponse(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
	type Uid struct {
		Sign            string `json:"sign"`
	}

	ar.log.Infof("处理请求响应 -> %s", path(ctx.Req))

	var remote string
	if resp.Request != nil {
		s := strings.Split(resp.Request.RemoteAddr,":")
		remote = s[0]
	}else{
		return resp
	}

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
		info := SignInfo{
			sign: uid.Sign,
			time: time.Now().Unix(),
		}
		ar.sign[remote] = info
		ar.log.Infof("解析Uid成功")
		resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		return resp
	} else if strings.HasSuffix(ctx.Req.URL.Path,"/Index/index"){
		sign := ar.sign[remote].sign
		data, err := cipher.AuthCodeDecodeB64(string(body)[1:], sign, true)
		if err != nil {
			ar.log.Errorf("解析用户数据失败 -> %+v", err)
			resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))
			return resp
		}
		ar.log.Infof("解析用户数据成功")
		body = []byte(regexp.MustCompile(ar.pattern).ReplaceAll([]byte(data), []byte(ar.replacement)))
		tmp,err := cipher.AuthCodeEncodeB64(string(body),sign)
		body = []byte("#" + tmp)
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
		//ar.log.Infof("请求 -> %s", path(req))
		if strings.HasSuffix(req.Host, "ppgame.com") {
			if strings.HasSuffix(req.URL.Path, "/Index/index") || strings.HasSuffix(req.URL.Path, "/Index/getDigitalSkyNbUid"){
				//ar.log.Infof("请求通过 -> %s", path(req))
				return true
			}
		}
		return false
	}
}

func (ar *AntiRivercrab) block() goproxy.ReqConditionFunc {
	return func(req *http.Request, ctx *goproxy.ProxyCtx) bool {
		if strings.HasSuffix(req.Host, "ppgame.com") {
			return false
		}else{
			//ar.log.Infof("请求拒绝 -> %s", path(req))
			return true
		}
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
