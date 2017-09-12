# SMARTPING #

## 介绍 ##

SmartPing为一个各机器(点)间间互PING检测工具，支持互PING，单向PING，绘制拓扑及报警功能

##功能 ##

 - 机器间互PING，单向PING，绘图
 - 绘制互PING间机器的拓扑及报警功能

## 设计思虑 ##
本系统设计为无中心化原则，所有的数据均存储自身点中，默认数据循环保留1个月时间，由自身点的数据绘制 出PING包 的状态，由各其他点的数据绘制 进PING包 的状态，并API接口获取其他点数据绘制整体PING拓扑图，拓扑图中存在报警功能，报警规则为Thdchecksec秒钟内发现大于等于Thdoccnum次延迟超过Thdavgdelay毫秒或丢包率大于Thdloss%即报警，若设置报警声音则会同时有Alertsound声音提醒。
    
## 配置文件 ##
    {
      "Name": "Localhost",
      "Ip": "127.0.0.1",
  
      "_comment_thd":"Thdchecksec,Thdoccnum,Thdloss,Thdavgdelay参数为设置报警规则，分别对应检测周期，丢包率，平均延迟，发生次数",
      "Thdchecksec" : 900,"Thdloss" : 30,"Thdavgdelay" : 200,"Thdoccnum" : 1,
    
      "_comment_topo":"Alertsound,Tline,Tsymbolsize,Topotimeout参数为设置拓扑报警时使用，分别对应报警声音，连线宽度，图形大小，拓扑间调用超时时间",
      "Alertsound" : "http://mp3.13400.com:99/1917/001204170042283.mp3","Tline":"1","Tsymbolsize":"70","Topotimeout":"5",
    
      "Targets": [
        {
          "Name": "本机","Addr": "127.0.0.1","Interval": "5","Type":"CS"
        },
        {
          "Name": "NOPING","Addr": "93.46.8.89","Interval": "5","Type":"CS"
        },
        
        {
          "Name": "IP_10_0_0_51","Addr": "10.0.0.51","Interval": "5","Type":"CS"
        },
        {
          "_comment":"报警规则（Thdchecksec,Thdoccnum,Thdloss,Thdavgdelay）可以单独设置，单独设置的规则优先",
          "Name": "IP_10_0_15_1","Addr": "10.0.15.1","Interval": "5","Type":"CS","Thdchecksec" : 30,"Thdoccnum" : 1,"Thdloss" : 50,"Thdavgdelay" : 200
        }
        {more...}
      ]
    }
    
**注意：**

  - 目标机器列表(Targets)中也需要包含本机的信息，否在在拓扑图中将不显示本机
  - PING检测包数需要小于60，目前绘图是一分钟一个点，若高于60将出现断点
  - 目标机器模式,C代表只接受PING包，不会发送PING包，CS及既发PING包也收PING包，拓扑图中体现为单项连线或双向连线
  - 若不设置Targets内的报警周期，检测次数，丢包率，平均延迟将采用基础配置


## 数据库 ##
数据库中存储PING的日志数据，默认循环保留一个月数据，最多（31*24*60*n=44640*N条）

    CREATE TABLE pinglog (
        logtime   VARCHAR (8),
        ip        VARCHAR (15),
        name      VARCHAR (15),
        maxdelay  VARCHAR (3),
        mindelay  VARCHAR (3),
        avgdelay  VARCHAR (3),
        sendpk    VARCHAR (2),
        revcpk    VARCHAR (2),
        losspk    VARCHAR (3),
        lastcheck VARCHAR (16),
        PRIMARY KEY (
            logtime,
            ip
        )
    );
    CREATE INDEX pk_n_t ON pinglog (
        name,
        lastcheck
    );
    CREATE INDEX pk_i_t ON pinglog (
        ip,
        lastcheck
    );


## 使用 ##

**运行**

./control start

**访问**

http://ip:8899

**首页示例：**

![index.jpg](_screen/index.jpg "")

绿色面积使用左侧坐标轴代表平均延迟，红色面积使用右侧坐标轴代表丢包率

**拓扑示例：**

![topology.jpg](_screen/topology.jpg "")

圆形代表一个点，绿色表示正常，红色表示当前查看拓扑图所在的点到此点不通，连线中绿色代表正常，红色代表异常，黄色代表未知(无法取到数据)

