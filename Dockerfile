FROM golang:1.16 as build
ADD go.mod /jarvis/go.mod
ADD go.sum /jarvis/go.sum
ADD cmd /jarvis/cmd
ADD models /jarvis/models
ADD vendor /jarvis/vendor
ADD restapi /jarvis/restapi
WORKDIR /jarvis
RUN go build -o jarvis cmd/jarvis-server/main.go

FROM alpine:latest as run
COPY --from=build /jarvis/jarvis /jarvis
ENTRYPOINT ["/jarvis"]