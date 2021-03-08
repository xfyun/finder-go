#include<stdio.h>
#include "config_center.h"
#include "libfinder.h"

int main(){

    char* project = "guiderAllService";
    char* group = "gas";
    char* service = "xist-ed";
    char* version = "1.0.0";
    char* file = "ist.toml";
    InitCenter("http://10.1.87.70:6868","10.1.87.43:33223");
    SubscribeConfigResult cfg = SubscribeFile(project,group,service,version,file);
    if (cfg.code != 0){
        printf("subscribe file error:%s",cfg.info);
        return 1;
    }
    printf("cfg:%s\n",cfg.data);
    for (;;){
        cfg = ListenFile(project,group,service,version,file,1);
        if (cfg.code != 0){
            printf("listen file error:%d,%s",cfg.code,cfg.info);
            break;
        }
        printf("cfg changed:->%s",cfg.data);
    }
}
