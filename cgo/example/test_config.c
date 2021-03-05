#include<stdio.h>
#include "config_center.h"
#include "libfinder.h"
int main(){
    InitCenter("http://10.1.87.70:6868","10.1.87.43:33223");
    SubscribeConfigResult cfg = SubscribeFile("guiderAllService", "gas","xist-ed", "1.0.0","ist.toml");
    if (cfg.code != 0){
        printf("subscribe file error:%s",cfg.info);
        return 1;
    }
    printf("cfg:%s\n",cfg.data);
    for (;;){
        cfg = ListenFile("guiderAllService", "gas","xist-ed", "1.0.0","ist.toml",1);
        if (cfg.code != 0){
            printf("listen file error:%d,%s",cfg.code,cfg.info);
            break;
        }
        printf("cfg changed:->%s",cfg.data);
    }
}
