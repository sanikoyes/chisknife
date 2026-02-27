export GOOS=windows
export GOARCH=amd64
export CGO_ENABLED=1
export CC=x86_64-w64-mingw32-gcc
export CXX=x86_64-w64-mingw32-g++
export HOST=x86_64-w64-mingw32

# go build -ldflags "-s -w -H=windowsgui -extldflags=-static" -p 4 -v
go build -ldflags "-s -extldflags=-static" -p 4 -v
