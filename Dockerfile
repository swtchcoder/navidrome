FROM --platform=$BUILDPLATFORM crazymax/osxcross:14.5-debian AS osxcross
FROM --platform=$BUILDPLATFORM tonistiigi/xx:1.5.0 AS xx

#####################################################
### Get TagLib
FROM --platform=$BUILDPLATFORM public.ecr.aws/docker/library/alpine:3.20 AS taglib-build
ARG TARGETPLATFORM
ARG CROSS_TAGLIB_VERSION=2.0.2-1
ENV CROSS_TAGLIB_RELEASES_URL=https://github.com/navidrome/cross-taglib/releases/download/v${CROSS_TAGLIB_VERSION}/

RUN PLATFORM=$(echo ${TARGETPLATFORM} | tr '/' '-') \
    FILE=taglib-${PLATFORM}.tar.gz && \
    DOWNLOAD_URL=${CROSS_TAGLIB_RELEASES_URL}${FILE} && \
    wget ${DOWNLOAD_URL}; \
    mkdir /taglib && \
    tar -xzf ${FILE} -C /taglib

#####################################################
### Build Navidrome UI
FROM --platform=$BUILDPLATFORM public.ecr.aws/docker/library/node:lts-alpine3.20 AS ui
WORKDIR /app

# Install node dependencies
COPY ui/package.json ui/package-lock.json ./
RUN npm ci

# Build bundle
COPY ui/ ./
RUN npm run build -- --outDir=/build

FROM scratch AS ui-bundle
COPY --from=ui /build /build

#####################################################
### Build Navidrome binary
FROM --platform=$BUILDPLATFORM public.ecr.aws/docker/library/golang:1.23-bookworm AS base
RUN apt-get update && apt-get install -y clang lld
COPY --from=xx / /
WORKDIR /workspace

FROM --platform=$BUILDPLATFORM base AS build

# Install build dependencies for the target platform
ARG TARGETPLATFORM
ARG GIT_SHA
ARG GIT_TAG

RUN xx-apt install -y binutils gcc g++ libc6-dev zlib1g-dev
RUN xx-verify --setup

RUN --mount=type=bind,source=. \
    --mount=type=cache,target=/root/.cache \
    --mount=type=cache,target=/go/pkg/mod \
    go mod download

RUN --mount=type=bind,source=. \
    --mount=from=ui,source=/build,target=./ui/build,ro \
    --mount=from=osxcross,src=/osxcross/SDK,target=/xx-sdk,ro \
    --mount=type=cache,target=/root/.cache \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=from=taglib-build,target=/taglib,src=/taglib,ro <<EOT


    # Setup CGO cross-compilation environment
    xx-go --wrap
    export CGO_ENABLED=1
    export PKG_CONFIG_PATH=/taglib/lib/pkgconfig
    cat $(go env GOENV)

    # Only Darwin (macOS) requires clang (default), Windows requires gcc, everything else can use any compiler.
    # So let's use gcc for everything except Darwin.
    if [ "$(xx-info os)" != "darwin" ]; then
        export CC=$(xx-info)-gcc
        export CXX=$(xx-info)-g++
        export LD_EXTRA="-extldflags '-static -latomic'"
    fi
    if [ "$(xx-info os)" = "windows" ]; then
        export EXT=".exe"
    fi

    go build -tags=netgo -ldflags="${LD_EXTRA} -w -s \
        -X github.com/navidrome/navidrome/consts.gitSha=${GIT_SHA} \
        -X github.com/navidrome/navidrome/consts.gitTag=${GIT_TAG}" \
        -o /out/navidrome${EXT} .
EOT

# Verify if the binary was built for the correct platform and it is statically linked
RUN xx-verify --static /out/navidrome*

FROM scratch AS binary
COPY --from=build /out /

#####################################################
### Build Final Image
FROM --platform=$BUILDPLATFORM public.ecr.aws/docker/library/alpine:3.20 AS final
LABEL maintainer="deluan@navidrome.org"

# Install ffmpeg and mpv
RUN apk add -U --no-cache ffmpeg mpv

# Copy navidrome binary
COPY --from=build /out/navidrome /app/

VOLUME ["/data", "/music"]
ENV ND_MUSICFOLDER=/music
ENV ND_DATAFOLDER=/data
ENV ND_PORT=4533
ENV GODEBUG="asyncpreemptoff=1"

EXPOSE ${ND_PORT}
HEALTHCHECK CMD wget -O- http://localhost:${ND_PORT}/ping || exit 1
WORKDIR /app

ENTRYPOINT ["/app/navidrome"]

