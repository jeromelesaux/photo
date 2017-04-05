install libexif on macox with brew : 
export CGO_CFLAGS=-I$(brew --prefix libexif)/include
export CGO_LDFLAGS=-L$(brew --prefix libexif)/lib
go get github.com/xiam/exif