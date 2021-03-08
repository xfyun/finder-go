#include<stdio.h>
#include "config_center.h"
#include "libfinder.h"

void printfList(SubscribeServiceResult* res ){
    int i;
    printf("addr is:");
    Node* list = res->addrList;
    for( i=0;i<res->length;i++){
        printf(" %s ",list->addr);
        list = list->next;
    }
    printf("\n");
}

int main(){
    InitCenter("http://10.1.87.70:6868","10.1.87.43:33223");
    SubscribeServiceResult* res = SubscribeService("guiderAllService", "gas", "myservice","webgate-ws", "1.0.0");
    if (res->code != 0){
        printf("subscribe service error :%s",res->info);
        return 0;
    }

    printfList(res);


    for (;;){
        res = ListenService("guiderAllService", "gas","webgate-ws", "1.0.0",1);
        if (res->code != 0){
                printf("subscribe service error :%s",res->info);
                return 0;
        }
        printf("service address  changed:->");
        printfList(res);
    }

}
