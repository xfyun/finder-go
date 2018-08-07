
#### 下拉镜像
````
docker pull 172.16.59.153/aiaas/findergo-demo:2.0.0
````
#### 创建以下目录
````
mkdir -p /opt/finder
````
#### 在finder目录下创建配置文件 config.cfg如下,请自行修改companionUrl 和 address以及group信息

/opt/finder/conf.json:
````
{
	"companionUrl": "http://10.1.87.70:6868",
	"address": "127.0.0.1:10010",
	"project": "qq",
	"group": "qq",
	"service": "qq",
	"version": "2.0",
	"subscribeFile": ["11.toml", "test2.yml"]
}

````
#### 创建start.sh
````
 sudo docker run --name findergo-test -v /home/yangzhou10/findergo-test:/root/go/src/git.xfyun.cn/AIaaS/finder-go/bin 172.16.59.153/aiaas/findergo-demo:2.0.0 ./demo /root/go/src/git.xfyun.cn/AIaaS/finder-go/bin/conf.json

````

#### 如果一台机器上需要启动多个客户端，创建多个config.cfg文件，放在不同的目录挂载上去,address也要不一样。
````
/opt/finder/conf.json
/opt/finder/conf1.json
/opt/finder/conf2.json
````
start1.sh:

````
sudo docker run --name findergo-test -v /home/yangzhou10/findergo-test:/root/go/src/git.xfyun.cn/AIaaS/finder-go/bin 172.16.59.153/aiaas/findergo-demo:2.0.0 ./demo /root/go/src/git.xfyun.cn/AIaaS/finder-go/bin/conf.json

````
start2.sh
````
sudo docker run --name findergo-test1 -v /home/yangzhou10/findergo-test:/root/go/src/git.xfyun.cn/AIaaS/finder-go/bin 172.16.59.153/aiaas/findergo-demo:2.0.0 ./demo /root/go/src/git.xfyun.cn/AIaaS/finder-go/bin/conf1.json
````
