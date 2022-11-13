# Stage 1: compile the program
FROM golang:1.19 as build-stage
WORKDIR /app
COPY go.* .
COPY cmd/ cmd/
COPY docs/ docs/
COPY internal/ internal/
COPY Makefile Makefile
RUN make build

# Stage 2: build the image
FROM alpine:latest  
RUN apk --no-cache add ca-certificates libc6-compat
COPY data/ data/
COPY --from=build-stage /app/bin/back .
EXPOSE 8080
CMD ["./back"]  