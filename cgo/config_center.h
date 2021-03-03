
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

#endif
