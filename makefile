CC=go
RM=rm
MV=mv


SOURCEDIR=.
SOURCES := $(shell find $(SOURCEDIR) -name '*.go' | grep -v '/vendor/')
#GOPATH=$(SOURCEDIR)/
#GOOS=linux
#GOARCH=amd64
#GOARCH=arm
#GOARM=7

EXEC=photo
EXEC1=photoexif
EXEC2=photocontroller


VERSION=1
BUILDDATE=$(shell date -u '+%s')
BUILDHASH=$(shell git rev-parse --short HEAD)


BUILD_TIME=`date +%FT%T%z`
PACKAGES := github.com/xiam/exif github.com/HouzuoGuo/tiedot/db  github.com/pkg/errors  github.com/disintegration/imaging  github.com/Sirupsen/logrus github.com/bshuster-repo/logrus-logstash-hook


LIBS= 

LDFLAGS=-ldflags "-s -X main.Version=$(VERSION) -X main.GitHash=$(BUILDHASH) -X main.BuildStmp=$(BUILDDATE)"

.DEFAULT_GOAL:= $(EXEC2)


$(EXEC2): organize $(SOURCES)  ${EXEC1}
		@echo "    Compilation des sources ${BUILD_TIME}"
		@if  [ "arm" = "${GOARCH}" ]; then\
		    	GOPATH=$(PWD)/../.. GOOS=${GOOS} GOARCH=${GOARCH} GOARM=${GOARM} go build ${LDFLAGS} -o ${EXEC2}-${VERSION} $(SOURCEDIR)/photocontroller/photocontroller.go;\
		else\
       			GOPATH=$(PWD)/../.. GOOS=${GOOS} GOARCH=${GOARCH} go build ${LDFLAGS} -o ${EXEC2}-${VERSION} $(SOURCEDIR)/photocontroller/photocontroller.go;\
        fi
		@echo "    ${EXEC2}-${VERSION} generated."


$(EXEC1): organize $(SOURCES)
		@echo "    Compilation des sources ${BUILD_TIME}"
		@if  [ "arm" = "${GOARCH}" ]; then\
		    GOPATH=$(PWD)/../.. GOOS=${GOOS} GOARCH=${GOARCH} GOARM=${GOARM} go build ${LDFLAGS} -o ${EXEC1}-${VERSION} $(SOURCEDIR)/photoexif/photoexif.go;\
		else\
	            GOPATH=$(PWD)/../.. GOOS=${GOOS} GOARCH=${GOARCH}  go build ${LDFLAGS} -o ${EXEC1}-${VERSION} $(SOURCEDIR)/photoexif/photoexif.go;\
        fi
		@echo "    ${EXEC1}-${VERSION} generated."

test: deps
		@go test -v $(shell go list ./... | grep -v '/vendor/')
		@echo " Tests OK."

deps: init
		@echo "    Download packages"
		@$(foreach element,$(PACKAGES),go get -d -v $(element);)

organize: deps
		@echo "    Go FMT"
		@$(foreach element,$(SOURCES),go fmt $(element);)

init: clean
		@echo "    Init of the project"
		@echo "    We compile for OS ${GOOS} and architecture ${GOARCH} and compiler $(shell go version)"

execute:
		./${EXEC1}-${VERSION}  -httpport 3001 -masteruri http://localhost:3000/register 2> photoexif.log &
		./${EXEC2}-${VERSION}  -configurationfile confclient.json -httpport 3000  2> photocontroller.log &

kill:
		$(shell killall -v photoexif-1)
		$(shell killall -v photocontroller-1)
		@echo "    Processes killed."

clean:
		@if [ -f "${EXEC1}-${VERSION}" ] ; then rm ${EXEC1}-${VERSION} ; fi
		@if [ -f "${EXEC2}-${VERSION}" ] ; then rm ${EXEC2}-${VERSION} ; fi
		@rm -fr database_photo.db
		@rm -f *.log
		@echo "    Nettoyage effectuee"

package:  ${EXEC1} ${EXEC2} swagger
		@zip -r ${EXEC}-${GOOS}-${GOARCH}-${VERSION}.zip ./${EXEC1}-${VERSION} ./${EXEC2}-${VERSION} resources
		@echo "    Archive ${EXEC}-${GOOS}-${GOARCH}-${VERSION}.zip created"

audit:   ${EXEC1}
		@go tool vet -all -shadow  $(shell go list ./... | grep -v /vendor/ | sed 's/photo\///')
		@golint $(shell go list ./... | grep -v /vendor/ | sed 's/photo\///')
		@echo "    Audit effectue"

swagger:
	@echo "Generate swagger json file specs"
	@GOPATH=$(PWD)/../.. GOOS=linux GOARCH=amd64 go run ${GOPATH}/src/github.com/go-swagger/go-swagger/cmd/swagger/swagger.go generate spec -m -b ./routes > resources/swagger.json
	@echo "Specs generate at resources/swagger.json"

#----------------------------------------------------------------------#
#----------------------------- docker actions -------------------------#
#----------------------------------------------------------------------#

#export DOCKER_HOST=$(shell ip route get 1 | awk '{if ($7 != "" ) { print "tcp://"$7":2375";}}')

DOCKER_IP=$(shell if [ -z "$(DOCKER_MACHINE_NAME)" ]; then echo 'localhost'; else docker-machine ip $(DOCKER_MACHINE_NAME); fi)

dockerBuild:
	docker build -t yula .

dockerClean:
	docker rmi -t yula .

dockerUp:
	docker-compose up -d

dockerStop:
	docker-compose stop
	docker-compose kill
	docker-compose rm -f

dockerBuildUp: dockerStop dockerBuild dockerUp

dockerWatch:
	@watch -n1 'docker ps | grep photo'

dockerLogs:
	docker-compose logs -f

