git clone -b develop-old https://github.com/xfyun/finder-go/v3.git

mkdir -p src/git.xfyun.cn/AIaaS

mv finder-go src/git.xfyun.cn/AIaaS

export GOPATH=$PWD

cd src/github.com/xfyun/finder-go/v3/cgo

sh build.sh

cd example && sh build.sh
