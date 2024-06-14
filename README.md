# uploadfile
uploadfile by qrcode
```
go install -a -ldflags="-s" github.com/dxasu/uploadfile@latest
```
提示: windows上, 关闭公用网络防火墙
-v 选项用于显示详细的构建信息。
-a 选项用于强制重新构建所有的包，即使它们似乎是最新的。
-ldflags="-s" 用于设置链接标志，这里是去除符号信息的标志。
