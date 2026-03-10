export GOOS=windows
export GOARCH=amd64
export CGO_ENABLED=1
export CC=x86_64-w64-mingw32-gcc
export CXX=x86_64-w64-mingw32-g++
export HOST=x86_64-w64-mingw32

# 将 icon.png 转换为 icon.ico
PATH_ICON_PNG=asset/icon/icon.png
PATH_ICON_ICO=asset/icon/icon.ico

if [ -f $PATH_ICON_PNG ]; then
    echo "正在将 $PATH_ICON_PNG 转换为 $PATH_ICON_ICO..."
    convert $PATH_ICON_PNG -define icon:auto-resize=256,128,64,48,32,16 $PATH_ICON_ICO
    if [ $? -eq 0 ]; then
        echo "图标转换成功"
    else
        echo "警告: 图标转换失败，请确保已安装 ImageMagick"
    fi
fi

# 编译 Windows 资源文件 (图标)
x86_64-w64-mingw32-windres asset/icon/icon.rc -O coff -o icon.syso

go build -ldflags "-s -w -H=windowsgui -extldflags=-static" -p 4 -v
# go build -ldflags "-s -extldflags=-static" -p 4 -v
