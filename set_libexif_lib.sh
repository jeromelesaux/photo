#! /bin/bash

export CGO_CFLAGS=-I$(brew --prefix libexif)/include
export CGO_LDFLAGS=-L$(brew --prefix libexif)/lib
