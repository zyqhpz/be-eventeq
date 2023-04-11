FROM golang:latest

# Set the Current Working Directory inside the container
RUN mkdir /build
WORKDIR /build

# Copy everything from the current directory to the PWD(Present Working Directory) inside the container
COPY go.mod go.sum ./

RUN export GO111MODULE=on

# Download all the dependencies
RUN go mod download

# Copy the source from the current directory to the PWD(Present Working Directory) inside the container
COPY . .

# Build the Go app
RUN go build -o main .

# This container exposes port 8080 to the outside world
EXPOSE 8080

ENTRYPOINT ["/build/main"]
