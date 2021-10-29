package funcs

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type FeiShu_Bot struct {
	App_id     string `json:"app_id"`
	App_secret string `json:"app_secret"`
	Token      string
	Expire     int64
	Cre_time   int64
}

type Token_resp struct {
	Code                int    `json:"code"`
	Expire              int64  `json:"expire"`
	Msg                 string `json:"msg"`
	Tenant_Access_Token string `json:"tenant_access_token"`
}
type Chat_obj struct {
	Avatar  string `json:"avatar"`
	Chat_id string `json:"chat_id"`
	Name    string `json:"name"`
}

type Dt struct {
	Groups []Chat_obj `json:"groups"`
}

type Chat_resp struct {
	Code int    `json:"code"`
	Data Dt     `json:"data"`
	Msg  string `json:"msg"`
}

type zh_ch struct {
	Title   string                `json:"title"`
	Content [][]map[string]string `json:"content"`
}

type post struct {
	Zh_ch zh_ch `json:"zh_cn"`
}
type Content struct {
	Post post `json:"post"`
}
type AtStruct struct {
	AtMobile string `json:"atMobile"`
	IsAtAll  bool   `json:"isAtAll"`
}
type Msg struct {
	Msg_type string   `json:"msg_type"`
	Content  Content  `json:"content"`
	At       AtStruct `json:"at"`
}

func Post_r(api_url string, fb_byte []byte) ([]byte, error) {
	buffer := bytes.NewBuffer(fb_byte)
	request, err := http.NewRequest("POST", api_url, buffer)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json;charset=UTF-8") //添加请求头
	client := http.Client{}                                              //创建客户端
	resp, err := client.Do(request.WithContext(context.TODO()))          //发送请求
	if err != nil {
		return nil, err
	}
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return respBytes, nil
}

func getc(raw string, tag string) []map[string]string {
	text := "text"
	if tag == "at" {
		text = "user_id"
	}
	c := map[string]string{
		"tag": tag,
		text:  raw,
	}
	ctt1 := make([]map[string]string, 0)
	ctt1 = append(ctt1, c)
	return ctt1
}

func SendFeishu(webhook string, title string, content string, isAtAll bool) string {
	var msg Msg
	var atObj AtStruct
	atObj.AtMobile = ""
	atObj.IsAtAll = isAtAll

	msg.At = atObj
	msg.Msg_type = "post"

	ctt2 := make([][]map[string]string, 0)
	ctt2 = append(ctt2, getc(content, "text"))
	//for _, v := range content {
	//	ctt2 = append(ctt2, getc(v, "text"))
	//}

	msg.Content.Post.Zh_ch.Title = title
	msg.Content.Post.Zh_ch.Content = ctt2
	content_byte, _ := json.Marshal(msg)
	respBytes, err := Post_r(webhook, content_byte)
	if err != nil {
		// todo
		println("error: ", err)
	}
	return string(respBytes)
}
