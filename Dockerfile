FROM golang:1.6

ENV CRANE_PATH ${GOPATH}/src/github.com/michaelsauter/crane
ENV GOBIN ${GOPATH}/bin

RUN mkdir -p ${CRANE_PATH}

COPY main.go ${GOPATH}
COPY crane ${CRANE_PATH}/crane
COPY vendor ${GOPATH}/vendor

RUN go-wrapper download \
    && for GOOS in darwin linux windows; do \
          echo "Building $GOOS" ;\
          export GOOS=$GOOS ;\
          export GOARCH=amd64 ;\
          go build -o ${GOBIN}/crane-$GOOS-$GOARCH main.go ;\
       done;
