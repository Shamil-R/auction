FROM golang AS builder

RUN go get -d github.com/golang/dep/cmd/dep && \
    go install github.com/golang/dep/cmd/dep

WORKDIR $GOPATH/src/gitlab/nefco/auction

COPY Gopkg.toml Gopkg.lock ./

RUN dep ensure --vendor-only -v

COPY . ./

RUN CGO_ENABLED=0 go build -o /auction ./cli/auction



FROM node AS docs

RUN npm install -g redoc-cli mobx

WORKDIR /docs

COPY swagger.yaml ./

RUN redoc-cli bundle -o index.html --title "Auction API" swagger.yaml



FROM golang:alpine

COPY --from=builder /auction .

COPY --from=docs /docs ./docs

CMD [ "./auction" ]