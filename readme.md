<p align="center">
    <a href="http://smartping.org">
        <img src="http://smartping.org/logo.png">
    </a>
    <h3 align="center">SmartPing | 开源、高效、便捷的网络质量监控神器</h3>
    <p align="center">
        各机器(点)间互PING检测工具，支持正向PING绘制，反向PING绘制，互PING拓扑绘制及报警功能
        <br>
        <a href="http://smartping.org"><strong>-- Browse website --</strong></a>
        <br>
        <br>
        <a href="https://goreportcard.com/report/github.com/gy-games/smartping">
            <img src="https://goreportcard.com/badge/github.com/gy-games/smartping" >
        </a>
         <a href="https://github.com/gy-games/smartping/releases">
             <img src="https://img.shields.io/github/release/gy-games/smartping.svg" >
         </a>
         <a href="https://github.com/gy-games/smartping/blob/master/LICENSE">
             <img src="https://img.shields.io/hexpm/l/plug.svg" >
         </a>
    </p>    
</p>

## 功能 ##

- 正向PING，反向Ping绘图
- 互PING间机器的状态拓扑
- 自定义延迟、丢包阈值报警

## 设计思路 ##

本系统设计为无中心化原则，所有的数据均存储自身点中，默认每个Ping目标点的数据循环保留1个月时间，由自身点的数据绘制 **出PING包** 的状态，由各其他点的数据绘制 **进PING包** 的状态，从任意一点查询数据均会通过Ajax请求关联点的API接口获取其他点数据组装全部数据，绘制 出Ping曲线图，进Ping曲线图，网络互Ping拓扑图。并可以设置阈值进行报警，方便对网络质量的监控。

- [去中心化](https://docs.smartping.org/arch/decentralized.html)
- [数据结构](https://docs.smartping.org/arch/data.html)

## 项目截图 ##

![app-bg.jpg](http://smartping.org/assets/img/app-bg.png "")

## 技术交流

<a target="_blank" href="//shang.qq.com/wpa/qunwpa?idkey=dd689e43fd8ecfeb28bffc31d53cb058c6ea23263aa1a34fc032efaf91aae924"><img border="0" src="http://pub.idqqimg.com/wpa/images/group.png" alt="SmartPing" title="SmartPing"></a>

## 项目贡献

欢迎参与项目贡献！比如提交PR修复一个bug，或者新建 [Issue](https://github.com/gy-games/smartping/issues/) 讨论新特性或者变更。

## 其他资料 ##

- 官网： http://smartping.org
- 文档： https://docs.smartping.org
- - 下载安装：https://docs.smartping.org/install.html
- - API文档：https://docs.smartping.org/api.html
