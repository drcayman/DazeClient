# DazeClient（客户端）

一个用Golang编写的免费、多功能、高性能代理DazeProxy的客户端。

DazeClient属于Daze代理套件。Daze代理套件包括：

1. [DazeProxy](https://github.com/crabkun/DazeProxy)--Daze代理服务端  
2. [DazeClient](https://github.com/crabkun/DazeClient)--Daze代理客户端  
3. [DazeAdmin](https://github.com/crabkun/DazeAdmin)--DazeProxy的数据库简单管理工具  

## DazeClient能为你提供什么功能？

- TCP、UDP代理转发（IPv4/IPv6）  
- 本地HTTPS/SOCKS5代理  
- TAP虚拟网卡全局代理（后期更新）
- 数据传输加密  
- 数据传输伪装  
- 本地加密DNS
- 第三方控制接口
- 模块化（加密和伪装均为模块化，方便第三方开发）

## 对于普通用户

用来连接DazeProxy服务器,然后作为代理使用。
后期更新TAP虚拟网卡之后，可以做到真正的全局代理，支持TCP、UDP网游。

## 对于开发者

DazeClient内置受控功能，可以指定控制地址，DazeClient运行后自动连接到此地址，发送命令可以控制DazeClient的行为，比如更改配置、代理端口，查看日志，查看网速等。利用这一特性，开发者可以开发出自己想要的外壳程序。

加密和伪装方式均为模块化设计，并统一和公开了相关接口。第三方如果有更好的想法，可以按照公开的接口进行开发加密方式或者伪装方式。

## 加密和伪装

目前Daze代理套件自带的伪装方式有
- none：无伪装
- http：可伪装成HTTP GET或POST连接  
- tls_handshake：可伪装成TLS1.2连接  

目前Daze代理套件自带的加密方式有
- none：无加密
- keypair-rsa：服务端生成RSA密钥并发送公钥与客户端协商aes密钥，然后进行aes128位cfb模式加密  
- psk-aes-128-cfb：客户端与服务端利用约定好的预共享密钥进行aes128位cfb模式加密  
- psk-aes-256-cfb：客户端与服务端利用约定好的预共享密钥进行aes256位cfb模式加密  
- psk-rc4-md5：客户端与服务端利用约定好的预共享密钥进行rc4加密  

## 哪里下载？
由于某些不可描述的原因，暂停下载一段时间

## 相关教程（持续更新中）
[客户端配置文件详解](https://github.com/crabkun/DazeClient/wiki/%E5%AE%A2%E6%88%B7%E7%AB%AF%E9%85%8D%E7%BD%AE%E6%96%87%E4%BB%B6%E8%AF%A6%E8%A7%A3)

[各加密方式的详细解释与区别](https://github.com/crabkun/DazeClient/wiki/%E5%90%84%E5%8A%A0%E5%AF%86%E6%96%B9%E5%BC%8F%E7%9A%84%E8%AF%A6%E7%BB%86%E8%A7%A3%E9%87%8A%E4%B8%8E%E5%8C%BA%E5%88%AB)

[各伪装方式的详细解释与区别](https://github.com/crabkun/DazeClient/wiki/%E5%90%84%E4%BC%AA%E8%A3%85%E6%96%B9%E5%BC%8F%E7%9A%84%E8%AF%A6%E7%BB%86%E8%A7%A3%E9%87%8A%E4%B8%8E%E5%8C%BA%E5%88%AB)

[PAC文件使用方法](https://github.com/crabkun/DazeClient/wiki/PAC%E6%96%87%E4%BB%B6%E4%BD%BF%E7%94%A8%E6%96%B9%E6%B3%95)

[客户端加密与伪装的开发文档](https://github.com/crabkun/DazeClient/wiki/%E5%AE%A2%E6%88%B7%E7%AB%AF%E5%8A%A0%E5%AF%86%E4%B8%8E%E4%BC%AA%E8%A3%85%E7%9A%84%E5%BC%80%E5%8F%91%E6%96%87%E6%A1%A3)

[客户端控制接口的命令详解](https://github.com/crabkun/DazeClient/wiki/%E5%AE%A2%E6%88%B7%E7%AB%AF%E6%8E%A7%E5%88%B6%E6%8E%A5%E5%8F%A3%E7%9A%84%E5%91%BD%E4%BB%A4%E8%AF%A6%E8%A7%A3)

[各种常见的问题与答案](https://github.com/crabkun/DazeProxy/wiki/%E5%90%84%E7%A7%8D%E5%B8%B8%E8%A7%81%E7%9A%84%E9%97%AE%E9%A2%98%E4%B8%8E%E7%AD%94%E6%A1%88)

## 感谢（Thanks）
本项目借助了以下开源项目的力量才能完成，非常感谢以下项目以及其作者们！  
- Xorm：[https://github.com/go-xorm/xorm](https://github.com/go-xorm/xorm)  
- Go-MySQL-Driver：[https://github.com/go-sql-driver/mysql](https://github.com/go-sql-driver/mysql)  
- go-sqlite3：[https://github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)  
- socks5：[https://github.com/physacco/socks5](https://github.com/physacco/socks5)
- gotun2socks：[https://github.com/yinghuocho/gotun2socks](https://github.com/yinghuocho/gotun2socks)
## 开源协议
BSD 3-Clause License

## 声明
本软件仅供技术交流和游戏网络延迟加速，并非侵入或非法控制计算机信息系统的软件，严禁将本软件用于商业及非法用途，如软件使用者不能遵守此规定，请马上停止使用并删除，对于因用户使用本软件而造成任何不良后果，均由用户自行承担，软件作者不负任何责任。您下载或者使用本软件，就代表您已经接受此声明，如产生法律纠纷与本人无关。