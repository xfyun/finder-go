
#ifndef __config_center__
#define __config_center__

typedef struct Node{
	char* addr;
	struct Node* next;
}Node;

typedef struct{
	int code;
	char* info;
	int length;
	Node* addrList;
}SubscribeServiceResult;


typedef struct{
    int code;  // 错误码 0表示没有错误
    char* info; //错误信息
    char* data; //文件内容
    char* name; // 文件名称
}SubscribeConfigResult;

typedef struct{
    int code;
    char* info;
}CommonResult;

#endif



/*
// 初始化
//@companionUrl： companion 的地址
//@myAddress: 服务发现上报的地址
extern void InitCenter(char* companionUrl, char* myAddress);

//myService；服务订阅者的服务名， service: 订阅的服务名称
//订阅服务
extern SubscribeServiceResult* SubscribeService(char* project, char* group, char* myService, char* service, char* apiVersion);

// 注册服务,使用初始化时传入的地址
extern CommonResult RegisterService(char* project, char* group, char* myService, char* apiVersion);

// 注册服务
extern CommonResult RegisterServiceWithAddr(char* project, char* group, char* myService, char* apiVersion,char* addr);

// 下线服务
extern CommonResult UnRegisterService(char* project, char* group, char* myService, char* apiVersion);

// 下线服务,指定下线的地址
extern CommonResult UnRegisterServiceWithAddr(char* project, char* group, char* myService, char* apiVersion,char* addr);

// 订阅配置文件
extern SubscribeConfigResult SubscribeFile(char* project, char* group, char* service, char* version, char* file);

// 监听服务,必须要先调用 SubscribeService 订阅要监听的服务,
// 当服务实例变化时，会返回所有的最新实例地址，没有变化时则会阻塞
// @ queue ：监听的队列，同一个线程监听的队列应该是一样的。
extern SubscribeServiceResult* ListenService(char* project, char* group, char* service, char* apiVersion,int queue);

// 监听配置 ，必须要先调用 SubscribeFile 订阅要监听的配置
// 当配置文件变更时，返回最新的配置文件，否则阻塞
//@ queue ：监听的队列，同一个线程监听的队列应该是一样的。
extern SubscribeConfigResult ListenFile(char* project, char* group, char* service, char* apiVersion,int queue);
*/
