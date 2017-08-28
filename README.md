## 如何修改这个项目的代码

### 安装docker

在mac下安装docker请 [参考这里](https://docs.docker.com/docker-for-mac/install/)

### 安装vscode

```
brew cask install visual-studio-code
```

### 安装golang
```
brew install go
```

### 编译并运行本项目
```
docker build -t html2image .
```

```
docker run -ti --rm -p 10070:8080 html2image
```

### 访问本服务

```
curl -v "http://127.0.0.1:8080/render?width=620\&html=abcdefg"
```

### TODO List

- [ ] 增加js判断图片是否已经全部加载完成
- [ ] 对输入的宽度width设定一个允许的范围
- [ ] 可以指定高度height并且设置一个允许的范围
- [ ] 可以换端口(不限于8080)
- [ ] 更好一点的超时处理机制
