#include<stdio.h>
#include "config_center.h"
#include "libfinder.h"


int main(){
    Init("http://10.1.87.70:6868","10.1.87.43:33223");
    SubscribeServiceResult* res = SubscribeService("guiderAllService", "gas", "xist-ed","xist-ed", "1.0.0");
    if (res->code != 0){
        printf("subscribe service error :%s",res->info);
        return 0;
    }
    Node* list = res->addrList;
    for(int i=0;i<res->length;i++){
        printf("addr is:%s",list->addr);
        list = list->next;
    }
}
