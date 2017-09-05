# photo is a self hosted pictures and photos manager.

## general purpose 
* Install it on what ever you want (windows, macos x, linux, raspberrypi), and
    * manage your albums
    * produce pdf albums 
    * keywords to retrieve your photos.
    * Search by exif tags etc...

## installation dependencies 
__install libexif on macox with brew :__ 
 * export CGO_CFLAGS=-I$(brew --prefix libexif)/include
 * export CGO_LDFLAGS=-L$(brew --prefix libexif)/lib
 * go get github.com/xiam/exif

## tests sets 
__raw images files sets :__ 

* http://www.cs.tut.fi/~lasip/foi_wwwstorage/NikonD80D_RAW_SAMPLES.zip
* http://www.cs.tut.fi/~lasip/foi_wwwstorage/Canon_EOS400D_RAW_SAMPLES.zip
* http://www.cs.tut.fi/~lasip/foi_wwwstorage/Canon_G10_RAW_SAMPLES.zip
