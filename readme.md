# socks-v5实现
* 支持用户名密码校验
* 支持tcp,udp代理
* 支持ipv6,ipv4,域名代理

# 如何使用？
```shell
export http_proxy = "socks5://<username>:<password>@<ip>:<port>"
# export http_proxy = "socks5://admin:admin@127.0.0.1:18888"
```
```shell
curl --socks5 <username>:<password>@<ip>:<port> www.baidu.com 
# curl --socks5 admin:amdin@127.0.0.1:18888 www.baidu.com 
```