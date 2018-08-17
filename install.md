
#### 下拉镜像
````
docker pull 172.16.59.153/aiaas/findergo-demo:2.0.0
````
#### 创建以下目录
````
mkdir -p /opt/finder
````
#### 在finder目录下创建配置文件 conf.json如下,请自行修改companionUrl 和 address以及group信息

#### 配置type=3 ，然后修改address、service、version信息来注册多个服务，启动一个demo，代表一个服务提供者
#### 配置type=2 ,修改subribeServiceItem来订阅多个服务。
#### 配置type=1,修改subscribeFile来订阅不同文件
/opt/finder/conf.json:
````
{
	"type" :3,
	"companionUrl": "http://10.1.87.70:6868",
	"address": "127.0.0.1:10010",
	"project": "zy_test",
	"group": "zy_test",
	"service": "zy_test1",
	"version": "2.0",
	"subscribeFile": ["11.toml"],
	"subribeServiceItem" :[{"serviceName":"zy_test1","apiVersion":"1.0"},{"serviceName":"zy_test1","apiVersion":"2.0"}]
}


````
#### 创建start.sh
````
 sudo docker run --name findergo-test -v /opt/finder:/root/go/src/git.xfyun.cn/AIaaS/finder-go/bin 172.16.59.153/aiaas/findergo-demo:2.0.0 ./demo /root/go/src/git.xfyun.cn/AIaaS/finder-go/bin/conf.json

````

#### 如果一台机器上需要启动多个客户端，创建多个conf.json文件，放在不同的目录挂载上去,address也要不一样。
````
/opt/finder/conf.json
/opt/finder/conf1.json
/opt/finder/conf2.json
````
start1.sh:

````
sudo docker run --name findergo-test -v /opt/finder:/root/go/src/git.xfyun.cn/AIaaS/finder-go/bin 172.16.59.153/aiaas/findergo-demo:2.0.0 ./demo /root/go/src/git.xfyun.cn/AIaaS/finder-go/bin/conf.json

````
start2.sh
````
sudo docker run --name findergo-test1 -v /opt/finder:/root/go/src/git.xfyun.cn/AIaaS/finder-go/bin 172.16.59.153/aiaas/findergo-demo:2.0.0 ./demo /root/go/src/git.xfyun.cn/AIaaS/finder-go/bin/conf1.json
````
