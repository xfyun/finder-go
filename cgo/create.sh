git clone -b develop-old https://git.iflytek.com/AIaaS/finder-go/v3.git

mkdir -p src/git.xfyun.cn/AIaaS

mv finder-go src/git.xfyun.cn/AIaaS

export GOPATH=$PWD

cd src/git.iflytek.com/AIaaS/finder-go/v3/cgo

sh build.sh

cd example && sh build.sh
