FROM golang:1.16 as build
ADD go.mod /jarvis/go.mod
ADD go.sum /jarvis/go.sum
ADD cmd /jarvis/cmd
ADD models /jarvis/models
ADD vendor /jarvis/vendor
ADD restapi /jarvis/restapi
ADD pkg /jarvis/pkg
WORKDIR /jarvis

ARG SKAFFOLD_GO_GCFLAGS
RUN echo "Go gcflags: ${SKAFFOLD_GO_GCFLAGS}"
RUN go build -gcflags="${SKAFFOLD_GO_GCFLAGS}" -mod=readonly -v -o jarvis cmd/jarvis-server/main.go

# Now create separate deployment image
FROM gcr.io/distroless/base as run

ENV GOTRACEBACK=single
ENV PORT 8080
ENV HOST 0.0.0.0

#FROM alpine:latest as run
COPY --from=build /jarvis/jarvis /jarvis

ENTRYPOINT ["/jarvis"]