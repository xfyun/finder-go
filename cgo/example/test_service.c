#include<stdio.h>
#include "config_center.h"
#include "libfinder.h"


int main(){
    InitCenter("http://10.1.87.70:6868","10.1.87.43:33223");
    SubscribeServiceResult* res = SubscribeService("guiderAllService", "gas", "myservice","webgate-ws", "1.0.0");
    if (res->code != 0){
        printf("subscribe service error :%s",res->info);
        return 0;
    }

    Node* list = res->addrList;
    int i;
    for( i=0;i<res->length;i++){
        printf("addr is:%s\n",list->addr);
        list = list->next;
    }
    for (;;){
        res = ListenService("guiderAllService", "gas","webgate-ws", "1.0.0",1);
          if (res->code != 0){
                printf("subscribe service error :%s",res->info);
                return 0;
           }
        printf("service address  changed:->");
        Node* list = res->addrList;
            for(i=0;i<res->length;i++){
                printf("addr is:%s\n",list->addr);
                list = list->next;
            }
    }

}
