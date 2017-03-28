CC=go
RM=rm
MV=mv


SOURCEDIR=.
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')
#GOPATH=$(SOURCEDIR)/
GOOS=linux
GOARCH=amd64
#GOARCH=arm
GOARM=7

EXEC=photo
EXEC1=photoexif
EXEC2=photocontroller

VERSION=1
BUILD_TIME=`date +%FT%T%z`
PACKAGES := github.com/xiam/exif github.com/HouzuoGuo/tiedot/db


LIBS= 

LDFLAGS=	

.DEFAULT_GOAL:= $(EXEC2)


$(EXEC2): organize $(SOURCES)  ${EXEC1}
		@echo "    Compilation des sources ${BUILD_TIME}"
		@if  [ "arm" = "${GOARCH}" ]; then\
		    GOPATH=$(PWD)/../.. GOOS=${GOOS} GOARCH=${GOARCH} GOARM=${GOARM} go build ${LDFLAGS} -o ${EXEC2}-${VERSION} $(SOURCEDIR)/photocontroller/photocontroller.go;\
		else\
            GOPATH=$(PWD)/../.. GOOS=${GOOS} GOARCH=${GOARCH} GOARM=${GOARM} go build ${LDFLAGS} -o ${EXEC2}-${VERSION} $(SOURCEDIR)/photocontroller/photocontroller.go;\
        fi
		@echo "    ${EXEC2}-${VERSION} generated."


$(EXEC1): organize $(SOURCES)
		@echo "    Compilation des sources ${BUILD_TIME}"
		@if  [ "arm" = "${GOARCH}" ]; then\
		    GOPATH=$(PWD)/../.. GOOS=${GOOS} GOARCH=${GOARCH} GOARM=${GOARM} go build ${LDFLAGS} -o ${EXEC1}-${VERSION} $(SOURCEDIR)/photoexif/photoexif.go;\
		else\
            GOPATH=$(PWD)/../.. GOOS=${GOOS} GOARCH=${GOARCH} GOARM=${GOARM} go build ${LDFLAGS} -o ${EXEC1}-${VERSION} $(SOURCEDIR)/photoexif/photoexif.go;\
        fi
		@echo "    ${EXEC1}-${VERSION} generated."

deps: init
		@echo "    Download packages"
		@$(foreach element,$(PACKAGES),go get -d -v $(element);)

organize: deps
		@echo "    Go FMT"
		@$(foreach element,$(SOURCES),go fmt $(element);)

init: clean
		@echo "    Init of the project"

execute:
		./${EXEC1}-${VERSION}  -httpport 3000
		./${EXEC2}-${VERSION}  -httpport 3001

clean:
		@if [ -f "${EXEC1}-${VERSION}" ] ; then rm ${EXEC1}-${VERSION} ; fi
		@if [ -f "${EXEC2}-${VERSION}" ] ; then rm ${EXEC2}-${VERSION} ; fi
		@echo "    Nettoyage effectuee"

package:  ${EXEC1} ${EXEC2} swagger
		@zip -r ${EXEC}-${GOOS}-${GOARCH}-${VERSION}.zip ./${EXEC1}-${VERSION} ./${EXEC2}-${VERSION} resources
		@echo "    Archive ${EXEC}-${GOOS}-${GOARCH}-${VERSION}.zip created"

audit:   ${EXEC1}
		@go tool vet -all -shadow ./
		@echo "    Audit effectue"

swagger:
	@echo "Generate swagger json file specs"
	@GOPATH=$(PWD)/../.. GOOS=linux GOARCH=amd64 go run ${GOPATH}/src/github.com/go-swagger/go-swagger/cmd/swagger/swagger.go generate spec -m -b ./routes > resources/swagger.json
	@echo "Specs generate at resources/swagger.json"
