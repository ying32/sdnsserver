# simplednsserver
simple dns server

** 这个只是写给自己用下的，这里也只作备份 **  


本机使用将127.0.0.1添加到网络设置属性的dns列表中，建议为首选  

启动时会读取当前目录下的 config.json文件  

```javascript    
{
	"dnss": ["114.114.114.114"], // 服务器列表，json数组
	"enabledthirdDNS": false,  //为true时，表示为当domains没找到相关域名时，使用dnss中的dns服务器进行查询
	"domains": [{  // 域名及服务器ip列表，json数组
		"domain": "your.domain.com",
		"ip": "192.168.1.10"
	}]
}
```

#### 第三方  
>  github.com/miekg/dns  

****
