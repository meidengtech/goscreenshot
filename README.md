## 如何修改这个项目的代码

### 安装 docker

在 mac 下安装 docker 请 [参考这里](https://docs.docker.com/docker-for-mac/install/)

### 安装 vscode

```
brew cask install visual-studio-code
```

### 安装 golang

```
brew install go
```

### 编译并运行本项目

```
docker build -t html2image .
```

```
docker run -ti --restart=always -p 10070:8080 html2image
```

### 访问本服务

```
curl -v "http://127.0.0.1:8080/render?width=620\&html=abcdefg"
```

### TODO List

- [x] 增加 js 判断图片是否已经全部加载完成
- [ ] 对输入的宽度 width 设定一个允许的范围
- [ ] 可以指定高度 height 并且设置一个允许的范围
- [ ] 可以换端口(不限于 8080)
- [x] 更好一点的超时处理机制
