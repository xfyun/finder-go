git clone -b develop-old https://git.iflytek.com/AIaaS/finder-go.git

mkdir -p src/git.xfyun.cn/AIaaS

mv finder-go src/git.xfyun.cn/AIaaS

export GOPATH=$PWD

cd src/git.iflytek.com/AIaaS/finder-go/cgo

sh build.sh

cd example && sh build.sh
