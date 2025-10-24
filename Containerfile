FROM docker.io/library/golang:1.22-alpine AS build

RUN apk add --no-cache \
    build-base cmake ninja perl linux-headers \
    openssl-dev ca-certificates git wget tar bash

WORKDIR /tmp

RUN git clone --depth 1 https://github.com/open-quantum-safe/liboqs.git && \
    cd liboqs && mkdir build && cd build && \
    cmake -GNinja -DCMAKE_INSTALL_PREFIX=/usr/local -DOQS_DIST_BUILD=ON .. && \
    ninja install && cd /tmp && rm -rf liboqs

RUN git clone --depth 1 https://github.com/open-quantum-safe/oqs-provider.git && \
    cd oqs-provider && \
    cmake -GNinja -B_build -DOQS_PROVIDER_MD=OFF -DBUILD_TESTING=OFF -DCMAKE_INSTALL_PREFIX=/usr/local . && \
    cmake --build _build && \
    cmake --install _build && \
    cd /tmp && rm -rf oqs-provider

RUN mkdir -p /oqs-dist && cp -a /usr/local/. /oqs-dist/

FROM docker.io/library/golang:1.22-alpine AS app

RUN apk add --no-cache ca-certificates openssl

ENV OSSL_PROVIDER_PATH=/usr/local/lib/ossl-modules
ENV LD_LIBRARY_PATH=/usr/local/lib

COPY --from=build /oqs-dist /usr/local

WORKDIR /keygen

COPY go.mod ./
RUN go mod tidy
COPY . .

RUN CGO_ENABLED=0 go build -o keygen ./cmd/keygen

ENTRYPOINT ["./keygen"]
