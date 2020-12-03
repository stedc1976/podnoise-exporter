FROM golang:1.15.5

# Set necessary environmet variables needed for our image
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Move to working directory /build
WORKDIR /build

# Copy and download dependency using go mod
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy the code into the container
COPY . .

# Build the application
RUN go build -a -installsuffix cgo -o podnoise-exporter .

FROM alpine:3.7

# Set up a non-root user
RUN adduser -u 666 -G root -h /home/myapp -D myapp

WORKDIR /myapp

# copy script files in ./script
RUN mkdir -p scripts
COPY ./scripts/*.sh ./scripts/
RUN chmod -R g+rwx ./scripts

# Use an unprivileged user.
USER myapp

# copy executable file
COPY --from=0 /build/podnoise-exporter .

#ENTRYPOINT ["tail", "-f", "/dev/null"]

CMD ["./podnoise-exporter"]