#####################################################
### Build UI bundles
FROM node:12.14-alpine AS jsbuilder
WORKDIR /src
COPY ui/package.json ui/package-lock.json ./
RUN npm ci
COPY ui/ .
RUN npm run build


#####################################################
### Build executable
FROM golang:1.13-alpine AS gobuilder

# Download build tools
RUN mkdir -p /src/ui/build
RUN apk add -U --no-cache build-base git
RUN go get -u github.com/go-bindata/go-bindata/...

# Download project dependencies
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

# Copy source and UI bundle, build executable
COPY . .
COPY --from=jsbuilder /src/build/* /src/ui/build/
COPY --from=jsbuilder /src/build/static/css/* /src/ui/build/static/css/
COPY --from=jsbuilder /src/build/static/js/* /src/ui/build/static/js/
RUN rm -rf /src/build/css /src/build/js
RUN make buildall

# Download and unpack static ffmpeg
ARG FFMPEG_VERSION=4.1.4
ARG FFMPEG_URL=https://www.johnvansickle.com/ffmpeg/old-releases/ffmpeg-${FFMPEG_VERSION}-amd64-static.tar.xz
ADD ${FFMPEG_URL} /tmp/ffmpeg.tar.xz
RUN cd /tmp && tar xJf ffmpeg.tar.xz


#####################################################
### Build Final Image
FROM alpine
COPY --from=gobuilder /src/sonic-server /app/
COPY --from=gobuilder /tmp/ffmpeg*/ffmpeg /usr/bin/

VOLUME ["/data", "/music"]
ENV SONIC_DBPATH /data/cloudsonic.db
ENV SONIC_MUSICFOLDER /music
ENV SONIC_LOGLEVEL info
EXPOSE 4533

WORKDIR /app
CMD "/app/sonic-server"
