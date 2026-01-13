# uses a lightweight linux image that has go installed
FROM golang:1.25-alpine

# sets the current folder inside the container to /app
WORKDIR /app

# copies files from computer into the container's /app folder
COPY . .

# compiles go code into an executable file named "main"
RUN go build -o main .

# tells docker that container listens on port 8080
EXPOSE 8080

# the command that runs the app when container starts
CMD ["./main"]