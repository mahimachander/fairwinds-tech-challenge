# Use an official Golang runtime as a parent image
FROM golang:1.19-alpine

# Set the working directory to /fairwinds-tech-challenge
WORKDIR /fairwinds-tech-challenge

# Copy the current directory contents into the container at /app
COPY . /fairwinds-tech-challenge

# Build the Go application
RUN go build -o myapp

# Define the command to run when the container starts
CMD ["/fairwinds-tech-challenge/myapp"]