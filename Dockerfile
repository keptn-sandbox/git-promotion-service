# Use the official Golang image to create a build artifact.
# This is based on Debian and sets the GOPATH to /go.
FROM golang:1.17.9 as builder-base
WORKDIR /go/src/github.com/keptn/keptn/git-promotion-service
RUN go install gotest.tools/gotestsum@latest

COPY go.mod go.sum ./
RUN go mod download -x

COPY . .

RUN if [ ! -z "$debugBuild" ]; then export BUILDFLAGS='-gcflags "all=-N -l"'; fi
RUN gotestsum --no-color=false -- -race -coverprofile=coverage.txt -covermode=atomic -v ./...
RUN GOOS=linux go build -ldflags '-linkmode=external' $BUILDFLAGS -v -o git-promotion-service

FROM markuslackner/keptn-production-base:0.0.1 as production
ARG version=develop
# required for external tools to detect this as a go binary
ENV GOTRACEBACK=all
LABEL org.opencontainers.image.source="https://github.com/keptn/keptn" \
    org.opencontainers.image.url="https://keptn.sh" \
    org.opencontainers.image.title="Keptn Git Promotion Service" \
    org.opencontainers.image.vendor="Keptn" \
    org.opencontainers.image.documentation="https://keptn.sh/docs/" \
    org.opencontainers.image.licenses="Apache-2.0" \
    org.opencontainers.image.version="${version}"

# Copy the binary to the production image from the builder stage.
COPY --from=builder-base /go/src/github.com/keptn/keptn/git-promotion-service/git-promotion-service /git-promotion-service
EXPOSE 8080

# Run the web service on container startup.
CMD ["/git-promotion-service"]