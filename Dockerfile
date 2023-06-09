
# 根据golang版本选择 
FROM public-cn-beijing.cr.volces.com/public/golang:alpine as builder  

# 指定构建过程中的工作目录
WORKDIR /app
# 将当前目录(dockerfile所在目录)下所有文件都拷贝到工作目录下（.dockerignore中文件除外)
COPY . /app/

# 执行代码编译命令
# 1、指定GOPROXY, 加快依赖拉取速度
# 2、操作系统参数为linux
# 3、可执行程序所在设备的CPU架构amd64
# 编译后的二进制产物命名为main, 并存放在当前目录下。
RUN GOPROXY=https://goproxy.cn,direct GOOS=linux GOARCH=amd64 go build -o main .

# 采用抖音云基础镜像, 包含https证书, bash, tzdata等常用命令
FROM public-cn-beijing.cr.volces.com/public/dycloud-golang:alpine-3.17
WORKDIR /opt/application
COPY --from=builder /app /opt/application
USER root

# 写入run.sh
RUN echo -e '#!/usr/bin/env bash\ncd /opt/application/ && ./main \n' > /opt/application/run.sh

# 指定run.sh权限
Run chmod a+x run.sh

EXPOSE 8000
