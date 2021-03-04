
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

//myService；服务订阅者的服务名， service: 订阅的服务名称
//订阅服务
extern SubscribeServiceResult* SubscribeService(char* project, char* group, char* myService, char* service, char* apiVersion);
// 注册服务
extern CommonResult RegisterService(char* project, char* group, char* myService, char* apiVersion);
// 下线服务
extern CommonResult UnRegisterService(char* project, char* group, char* myService, char* apiVersion);
// 订阅配置文件
extern SubscribeConfigResult SubscribeFile(char* project, char* group, char* service, char* version, char* file);
// 初始化
extern void InitCenter(char* companionUrl, char* myAddress);
// 监听服务,必须要先调用 SubscribeService 订阅要监听的服务
extern SubscribeServiceResult* ListenService(char* project, char* group, char* service, char* apiVersion);
// 监听配置 ，必须要先调用 SubscribeFile 订阅要监听的配置
extern SubscribeConfigResult ListenFile(char* project, char* group, char* service, char* apiVersion);


*/
