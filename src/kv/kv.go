package kv

import (
	"crypto/md5"
	"fmt"
	"sync"

	"github.com/hashicorp/consul/api"
)

func getSession(client *api.Client, hostIP string) (id string, err error) {
	// 检查session id是否存在
	kv := client.KV()
	pair, _, err := kv.Get(fmt.Sprintf("smartping/session_ids/%s", hostIP), nil)
	if err != nil {
		return "", err
	}

	if pair == nil { // 不存在，则创建一个
		entry := &api.SessionEntry{Behavior: api.SessionBehaviorDelete, TTL: "120s"} // ttl为2分钟
		id, _, err = client.Session().Create(entry, nil)
		if err != nil {
			return
		} else {
			if _, _, err = kv.Acquire(&api.KVPair{
				Key:     fmt.Sprintf("smartping/session_ids/%s", hostIP),
				Value:   []byte(id),
				Session: id,
			}, nil); err != nil {
				return
			}
		}
	} else {
		id = string(pair.Value[:])
	}
	return
}

// Registry 注册到consul中
func Registry(consulEndpoint, hostName, hostIP string) (err error) {
	client, err := GetClient(consulEndpoint)
	if err != nil {
		return err
	}
	kv := client.KV()

	id, err := getSession(client, hostIP)
	if err != nil {
		return err
	}

	// 组装数据
	key := fmt.Sprintf("smartping/targets/hosts/%s", hostIP)
	value := []byte(fmt.Sprintf(`{
	"Name": "%s",
	"Addr": "%s"
}`, hostName, hostIP))

	// 获取旧的value数据
	var oldValue []byte
	if pair, _, err := kv.Get(key, nil); err != nil {
		return err
	} else if pair != nil {
		oldValue = pair.Value
	}

	// 对比是否需要更新, 设置相同数据可能造成consul-template重新渲染
	if fmt.Sprintf("%x", md5.Sum(value)) != fmt.Sprintf("%x", md5.Sum(oldValue)) {
		pair := &api.KVPair{
			Session: id,
			Key:     key,
			Value:   value,
		}
		if _, _, err = kv.Acquire(pair, nil); err != nil {
			return
		}
	}

	// 刷新id
	_, _, err = client.Session().Renew(id, nil)
	return err
}

// Delete 删除
func Delete(consulEndpoint, hostIP string) (err error) {
	client, err := GetClient(consulEndpoint)
	if err != nil {
		return err
	}
	pair, _, err := client.KV().Get(fmt.Sprintf("smartping/session_ids/%s", hostIP), nil)
	if err != nil {
		return err
	} else if pair != nil {
		_, err = client.Session().Destroy(string(pair.Value), nil)
	}
	return
}

var c *api.Client
var once sync.Once

// GetClient 获取客户端
func GetClient(consulEndpoint string) (client *api.Client, err error) {
	once.Do(func() {
		c, err = api.NewClient(&api.Config{Address: consulEndpoint, Scheme: "http"})
	})
	if err != nil {
		return
	}
	client = c
	return
}
