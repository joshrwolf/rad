FROM golang:1.13-alpine AS build

ENV CGO_ENABLED 0
ENV GOOS linux

# Fetch build dependencies
RUN apk add git make openssl

WORKDIR /app

# Copy src code
COPY . .

# Build the application
RUN make app

# Just Go things
FROM scratch

COPY --from=build /app/rad .
ENTRYPOINT [ "/rad" ]