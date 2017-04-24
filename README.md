__install libexif on macox with brew :__ 
 * export CGO_CFLAGS=-I$(brew --prefix libexif)/include
 * export CGO_LDFLAGS=-L$(brew --prefix libexif)/lib
 * go get github.com/xiam/exif

__raw images files sets :__ 

* http://www.cs.tut.fi/~lasip/foi_wwwstorage/NikonD80D_RAW_SAMPLES.zip
* http://www.cs.tut.fi/~lasip/foi_wwwstorage/Canon_EOS400D_RAW_SAMPLES.zip
* http://www.cs.tut.fi/~lasip/foi_wwwstorage/Canon_G10_RAW_SAMPLES.zip
