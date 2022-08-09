package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Task struct {
	Status string `json:"status"`
	Code   int    `json:"code"`
	Msg    string `json:"msg"`
	Data   []struct {
		Ip   string `json:"ip"`
		Sign string `json:"sign"`
	} `json:"data"`
}
type IpData struct {
	Ret  string   `json:"ret"`
	IP   string   `json:"ip"`
	Data []string `json:"data"`
}

var wg sync.WaitGroup

func main() {
	//发送httpGET
	client := &http.Client{}
	request, e := http.NewRequest("GET", "https://site.ip138.com/domain/read.do?domain=cloudnproxy.baidu.com&time="+string(rune(time.Microsecond)), nil)
	if e != nil {
		fmt.Println(e.Error())
		return
	}
	//header
	request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/80.0.3987.149 Safari/537.36")
	//发送请求
	resp, e := client.Do(request)
	if e != nil {
		fmt.Println(e.Error())
		return
	}
	defer resp.Body.Close()
	s, _ := ioutil.ReadAll(resp.Body)
	task := new(Task)
	_ = json.Unmarshal(s, &task)
	type SendData []struct {
		Ip       string `json:"ip"`
		Location string `json:"location"`
	}
	wg.Add(len(task.Data))
	var sendData SendData
	for _, v := range task.Data {

		go func(v struct {
			Ip   string `json:"ip"`
			Sign string `json:"sign"`
		}) {
			defer wg.Done()
			req, e := http.NewRequest("GET", "https://api.ip138.com/query/?ip="+v.Ip+"&oid=5&mid=5&datatype=json&sign="+v.Sign+"&callback=", nil)
			if e != nil {
				fmt.Println(e.Error())
				return
			}
			req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/80.0.3987.149 Safari/537.36")
			res, e := client.Do(req)
			if e != nil {
				fmt.Println(e.Error())
				return
			}
			defer res.Body.Close()
			s, _ := ioutil.ReadAll(res.Body)
			ipData := new(IpData)
			_ = json.Unmarshal(s, &ipData)
			var str string
			for i, value := range ipData.Data {
				if i < 4 {
					str += value
				}
			}
			sendData = append(sendData, struct {
				Ip       string `json:"ip"`
				Location string `json:"location"`
			}{
				Ip:       v.Ip,
				Location: str,
			})
		}(v)

	}
	wg.Wait()

	type Send struct {
		Code int      `json:"code"`
		Msg  string   `json:"msg"`
		Data SendData `json:"data"`
	}
	send := new(Send)
	send.Code = 200
	send.Msg = "获取成功"
	send.Data = sendData
	b, _ := json.Marshal(send)
	PushMI(string(b))
	//等待八小时
	wg.Done()

}

// PushMI 小米推送函数
func PushMI(m string) {
	// 设置小米Push推送URL
	var urlPush = "https://api.xmpush.xiaomi.com/v3/message/all"
	// 初始化http客户端
	client := &http.Client{}
	// 初始化url Body值
	data := url.Values{
		//UrlEncode(m)

		"payload": {url.QueryEscape(m)},
		// 消息的内容。（注意：需要对payload字符串做urlencode处理）
		"restricted_package_name": {"com.wusui.msg"},
		// App的包名。备注：V2版本支持一个包名，V3版本支持多包名（中间用逗号分割）。
		"pass_through": {"0"},
		// pass_through的值可以为： 0 表示通知栏消息 1 表示透传消息
		"title": {"每日更新免流IP"},
		// 通知栏展示的通知的标题，不允许全是空白字符，长度小于50， 一个中英文字符均计算为1（通知栏消息必填）。
		"description": {"速更"},
		// 通知栏展示的通知的描述，不允许全是空白字符，长度小于128，一个中英文字符均计算为1（通知栏消息必填）。
		"notify_type":  {"1"},
		"time_to_live": {"20000"}, // 可选项。如果用户离线，设置消息在服务器保存的时间，单位：ms。服务器默认最长保留两周。
	}
	// 发送Request请求 请求方式为POST
	req, e := http.NewRequest("POST", urlPush, strings.NewReader(data.Encode())) // 建立一个请求
	if e != nil {
		fmt.Println(e.Error())
		return
	}
	// 检查错误函数
	// request请求添加Headers
	req.Header.Add("Accept", "*/*")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:12.0) Gecko/20100101 Firefox/12.0")
	//更改Mipush服务器的认证信息
	req.Header.Add("Authorization", "key=VTisfjw1GU48XYdtlNn03w==") //
	req.Header.Add("content-type", "application/x-www-form-urlencoded")
	req.Header.Add("content-length", strconv.Itoa(len(data.Encode())))
	// 获取http响应
	res, _ := client.Do(req)
	//读取body
	// 代码执行完毕后，关闭Body
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println(err.Error())
		}
	}(res.Body)
}
