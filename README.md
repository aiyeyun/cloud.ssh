Welcome to Web Cloud SSH !
===================

GitHub https://github.com/aiyeyun/cloud.ssh.git 

演示DEMO http://ssh.aiyeyun.com/

Demo
-------------
![cloud.ssh.demo](https://raw.githubusercontent.com/aiyeyun/cloud.ssh/master/cloud.ssh.demo.gif "演示DEMO") 

Installation Guide
-------------
> **Note:**

> - 安装golang 环境
> - cd $GOPATH/src
> - git clone https://github.com/aiyeyun/cloud.ssh.git
> - go get golang.org/x/crypto/ssh [需翻墙下载 或 使用以下方式安装 ]
> - cd $GOPATH/src
> - mkdir golang.org
> - cd golang.org
> - mkdir x
> - cd x
> - git clone https://github.com/golang/crypto.git
> - go get github.com/go-ini/ini
> - go get github.com/gorilla/websocket
> - cd $GOPATH/src/cloud.ssh
> - go install
> - cd $GOPATH/bin/
> - ./cloud.ssh

Directory Structure
-------------
config

    --config.ini 配置端口
