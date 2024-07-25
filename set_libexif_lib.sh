#! /bin/bash

export CGO_CFLAGS=-I$(brew --prefix)/include
export CGO_LDFLAGS=-L$(brew --prefix)/lib
