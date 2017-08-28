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

