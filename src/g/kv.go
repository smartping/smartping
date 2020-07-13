package g

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cihub/seelog"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	"github.com/smartping/smartping/src/kv"
)

var version string

// StartAutoDiscoveryConfig4LocalMode 开启local模式配置自动更新
func StartAutoDiscoveryConfig4LocalMode(consulEndpoint string, ver string) {
	if hostName, hostIP, err := kv.GetHostInfo(); err == nil {
		Cfg.Addr = hostIP
		Cfg.Name = hostName
	} else {
		seelog.Errorf("StartAutoDiscoveryConfig4LocalMode ——> %v", err)
	}
	version = ver
	// 判断是否是cloud模式
	if Cfg.Mode["Type"] != "cloud" {
		pipe := make(chan AgentInfos, 1)
		go func() {
			AutoDiscoveryHostAgent(consulEndpoint, pipe)
		}()
		go func() {
			GenerateHostAgentConfig(pipe, UpdateAgentConfig)
		}()
	}
}

// AutoDiscoveryHostAgent consul中自动获取agent相关信息
func AutoDiscoveryHostAgent(consulEndpoint string, pipe chan<- AgentInfos) error {
	if pipe == nil {
		return fmt.Errorf("nil chan")
	}

	plan, err := watch.Parse(map[string]interface{}{
		"type":   "keyprefix",
		"prefix": "smartping/targets/",
	})
	if err != nil {
		return err
	}
	defer func() {
		plan.Stop()
		close(pipe)
	}()

	plan.Handler = func(idx uint64, raw interface{}) {
		if raw == nil {
			return // ignore
		}
		kvpairs, ok := raw.(api.KVPairs)

		if !ok || kvpairs == nil {
			return // ignore
		}
		agentInfos := AgentInfos{}
		// var items = make([]HostAgentInfo, 0)
		for _, kvpair := range kvpairs {
			seelog.Infof("receive agent info: %v", kvpair)
			if kvpair.Value == nil || len(kvpair.Value) == 0 {
				continue
			}
			var data HostAgentInfo
			if err := json.Unmarshal(kvpair.Value, &data); err != nil {
				seelog.Errorf("unmarshal byte data: %s json data error: %v", kvpair.Value, err)
				continue
			}

			if strings.Contains(kvpair.Key, "smartping/targets/hosts/") {
				if agentInfos.Agents == nil {
					agentInfos.Agents = make([]HostAgentInfo, 0)
				}
				agentInfos.Agents = append(agentInfos.Agents, data)
			} else if strings.Contains(kvpair.Key, "smartping/targets/slbs/") {
				if agentInfos.NoAgents == nil {
					agentInfos.NoAgents = make([]HostAgentInfo, 0)
				}
				agentInfos.NoAgents = append(agentInfos.NoAgents, data)
			}

		}
		pipe <- agentInfos
	}

	if err := plan.Run(consulEndpoint); err != nil {
		return err
	}
	return nil
}

// GenerateHostAgentConfig 生成agent配置
func GenerateHostAgentConfig(pipe <-chan AgentInfos, handler Handler) {
	for data := range pipe {
		if handler != nil {
			handler(data)
		}
	}
}

// Handler 数据处理函数
type Handler func(datas AgentInfos)

// DefaultHandler 默认数据处理函数
func DefaultHandler(datas AgentInfos) {
	seelog.Infof("==========>> %v", datas)
}

// UpdateAgentConfig 更新配置数据内容
func UpdateAgentConfig(datas AgentInfos) {
	if datas.checkUpdate(Cfg.Network) {
		SaveConfig()
		ParseConfig(version)
	}
}

// HostAgentInfo the info of agent registried
type HostAgentInfo struct {
	Name    string `json:"Name"`
	Address string `json:"Addr"`
}

// HostNoAgentInfos 未安装了agent的host
type HostNoAgentInfos []HostAgentInfo

// HostAgentInfos 安装了agent的host
type HostAgentInfos []HostAgentInfo

// AgentInfos agent信息
type AgentInfos struct {
	Agents   HostAgentInfos
	NoAgents HostNoAgentInfos
}

func (ai AgentInfos) toNetworkMembers() (members map[string]NetworkMember) {
	if ai.Agents != nil && len(ai.Agents) > 0 {
		for _, a := range ai.Agents {
			if members == nil {
				members = make(map[string]NetworkMember)
			}
			members[a.Address] = ai.toDefaultMember(a, true)
		}
	}

	if ai.NoAgents != nil && len(ai.NoAgents) > 0 {
		for _, a := range ai.NoAgents {
			if members == nil {
				members = make(map[string]NetworkMember)
			}
			members[a.Address] = ai.toDefaultMember(a, false)
		}
	}
	return
}

func (ai AgentInfos) checkUpdate(members map[string]NetworkMember) bool {
	current, err := json.Marshal(members)
	if err != nil {
		seelog.Warnf("Md5Sum when json.Marshal current data error: %v", err)
		return true
	}
	newMembers := ai.toNetworkMembers()
	newDatas, err := json.Marshal(newMembers)
	if err != nil {
		seelog.Warnf("Md5Sum when json.Marshal new data error: %v", err)
		return true
	}
	omd5 := fmt.Sprintf("%x", md5.Sum(current))
	nmd5 := fmt.Sprintf("%x", md5.Sum(newDatas))
	if omd5 != nmd5 {
		Cfg.Network = newMembers
		return true
	}

	return false
}

func (ai AgentInfos) toDefaultMember(agent HostAgentInfo, hasAgent bool) (nm NetworkMember) {
	nm = NetworkMember{
		Name:      agent.Name,
		Addr:      agent.Address,
		Ping:      []string{},
		Topology:  make([]map[string]string, 0),
		Smartping: false,
	}
	if hasAgent {
		nm.Smartping = true
		nm.Ping = ai.getPings4Agent(agent)
		nm.Topology = ai.getTopology4Agent(agent)
	}
	return
}

func (ai AgentInfos) getPings4Agent(agent HostAgentInfo) (pings []string) {
	pings = make([]string, 0)
	if ai.Agents != nil && len(ai.Agents) > 0 {
		for _, a := range ai.Agents {
			if a.Address != agent.Address {
				pings = append(pings, a.Address)
			}
		}
	}

	if ai.NoAgents != nil && len(ai.NoAgents) > 0 {
		for _, a := range ai.NoAgents {
			if a.Address != agent.Address {
				pings = append(pings, a.Address)
			}
		}
	}
	return
}

func (ai AgentInfos) getTopology4Agent(agent HostAgentInfo) (topologies []map[string]string) {
	topologies = make([]map[string]string, 0)
	if ai.Agents != nil && len(ai.Agents) > 0 {
		for _, a := range ai.Agents {
			if a.Address != agent.Address {
				topology := map[string]string{
					"Addr":        a.Address,
					"Name":        a.Name,
					"Thdavgdelay": "200",
					"Thdchecksec": "900",
					"Thdloss":     "30",
					"Thdoccnum":   "3",
				}
				topologies = append(topologies, topology)
			}
		}
	}

	if ai.NoAgents != nil && len(ai.NoAgents) > 0 {
		for _, a := range ai.NoAgents {
			if a.Address != agent.Address {
				topology := map[string]string{
					"Addr":        a.Address,
					"Name":        a.Name,
					"Thdavgdelay": "200",
					"Thdchecksec": "900",
					"Thdloss":     "30",
					"Thdoccnum":   "3",
				}
				topologies = append(topologies, topology)
			}
		}
	}
	return
}
