# Start from the latest golang base image
FROM golang:latest

# Add Maintainer Info
LABEL maintainer="Your Name <youremail@domain.com>"

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .


# Create .ssh directory and known_hosts file
RUN mkdir -p /root/.ssh && touch /root/.ssh/known_hosts
# Build the Go app
RUN make build


# Command to run the executable
CMD ["/main"]