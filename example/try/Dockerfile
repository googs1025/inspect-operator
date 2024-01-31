FROM golang:1.17


ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct
RUN mkdir /root/try
COPY . /root/try
WORKDIR /root/try

# CGO_ENABLED禁用cgo 然后指定OS等，并go build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build  -o try .

EXPOSE 8000
ENTRYPOINT ["./try"]


# 用来测试job的镜像使用