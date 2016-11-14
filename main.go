package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"runtime"

	"github.com/miekg/dns"
)

type TMyItem struct {
	Domain string `json:"domain"`
	IP     string `json:"ip"`
}

type TServerConfig struct {
	DNSS            []string  `json:"dnss"`
	EnabledThirdDNS bool      `json:"enabledthirdDNS"`
	Domains         []TMyItem `json:"domains"`
}

var (
	gSvrCfg    TServerConfig
	gDomainMap map[string]string
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU()*2 - 1)

	dns.HandleFunc(".", dnsHandle)

	local := ":53"

	failure := make(chan error, 1)

	go func(failure chan error) {
		failure <- dns.ListenAndServe(local, "tcp", nil)
	}(failure)

	go func(failure chan error) {
		failure <- dns.ListenAndServe(local, "udp", nil)
	}(failure)

	fmt.Printf("已准备接收来自%s的tcp/udp协议报文 ...\n", local)

	fmt.Println(<-failure)
}

func init() {
	bytes, err := ioutil.ReadFile("./config.json")
	if err != nil {
		fmt.Println("配置文件不存在！")
		return
	}
	err = json.Unmarshal(bytes, &gSvrCfg)
	if err != nil {
		fmt.Println("配置文件解析失败，消息：", err)
		return
	}
	gDomainMap = make(map[string]string, 1)
	for _, v := range gSvrCfg.Domains {
		k := v.Domain + "."
		if _, ok := gDomainMap[k]; !ok {
			gDomainMap[k] = v.IP
		}
	}
}

func getNetDnsResult(host string, req *dns.Msg) *dns.Msg {
	c := &dns.Client{}
	c.Net = "udp"
	mm, _, err := c.Exchange(req, host+":53")
	if err != nil {
		return nil
	}
	return mm
}

func dnsHandle(w dns.ResponseWriter, req *dns.Msg) {

	//fmt.Println(req)

	if req.MsgHdr.Response == true { // supposed responses sent to us are bogus
		return
	}
	// 测试返回的什么
	//	c := new(dns.Client)
	//	c.Net = "udp"
	//	mm, _, err := c.Exchange(req, "202.96.128.86:53")
	//	fmt.Println(err)
	//	n := len(mm.Answer)
	//	for i := 0; i < n; i++ {
	//		fmt.Println("第", i, "个：", mm.Answer[i].String())
	//		/*
	//		        <nil>
	//		第 0 个： www.baidu.com.	600	IN	CNAME	www.a.shifen.com.
	//		第 1 个： www.a.shifen.com.	600	IN	A	14.215.177.38
	//		第 2 个： www.a.shifen.com.	600	IN	A	14.215.177.37

	//		*/
	//	}
	//	return

	if len(req.Question) == 1 {
		if req.Question[0].Qtype == 1 { // A

			domain := req.Question[0].Name
			fmt.Println("收到A记录查询请求：", domain)
			if v, ok := gDomainMap[domain]; ok {
				m := req.Copy()
				m.Rcode = 0 // 0 - 无差错
				m.Answer = make([]dns.RR, 2)
				var err error
				m.Answer[0], err = dns.NewRR(fmt.Sprintf("%s\t600\tIN\tCNAME\t%s\t", domain, domain))
				m.Answer[1], err = dns.NewRR(fmt.Sprintf("%s\t600\tIN\tA\t%s\t", domain, v))
				err = w.WriteMsg(m)
				if err != nil {
					fmt.Println("回复消息失败！=", err)
				} else {
					fmt.Println("已经回“" + req.Question[0].Name + "”的复消息！")
				}
			} else {
				if gSvrCfg.EnabledThirdDNS {
					for _, v := range gSvrCfg.DNSS {
						m := getNetDnsResult(v, req)
						if m != nil {
							w.WriteMsg(m)
							return
						}
					}
				} else {
					fmt.Println("未到找:", req.Question[0].Name, ", 忽略本次请求！")
				}
			}
		}
	} else {
		fmt.Println("Question len = 0")
	}

}
