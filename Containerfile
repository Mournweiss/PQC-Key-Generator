# Build OpenSSL and OQS/Provider Layer
FROM docker.io/library/golang:1.25.3-alpine AS build

RUN apk add --no-cache \
    build-base cmake ninja perl linux-headers \
    openssl-dev ca-certificates git wget tar bash

# Download and build OpenSSL 3.5.0 from source
RUN apk add --no-cache build-base perl cmake ninja git wget bash tar linux-headers ca-certificates && \
    wget https://www.openssl.org/source/openssl-3.5.0.tar.gz && \
    tar xzf openssl-3.5.0.tar.gz && cd openssl-3.5.0 && \
    ./config --prefix=/usr/local/ssl --openssldir=/usr/local/ssl shared && \
    make -j$(nproc) && make install_sw && \
    cd /tmp && rm -rf openssl-3.5.0*

# Build liboqs (Open Quantum Safe) library
RUN git clone --branch main https://github.com/open-quantum-safe/liboqs.git && \
    cd liboqs && mkdir build && cd build && \
    cmake -GNinja -DCMAKE_INSTALL_PREFIX=/usr/local .. && \
    ninja install && cd /tmp && rm -rf liboqs

# Build OQS Provider for OpenSSL_3
RUN git clone --branch main https://github.com/open-quantum-safe/oqs-provider.git && \
    cd oqs-provider && mkdir build && cd build && \
    cmake -GNinja -DOPENSSL_ROOT_DIR=/usr/local/ssl -DCMAKE_INSTALL_PREFIX=/usr/local .. && \
    ninja install && cd /tmp && rm -rf oqs-provider

# Bundle results for next stage
RUN mkdir -p /oqs-dist && cp -a /usr/. /oqs-dist/

# Final Application Image
FROM docker.io/library/golang:1.25.3-alpine AS app

RUN apk add --no-cache ca-certificates openssl

ENV PATH="/usr/local/ssl/bin:$PATH"
ENV LD_LIBRARY_PATH="/usr/local/ssl/lib:/usr/local/lib"
ENV OSSL_PROVIDER_PATH="/usr/local/ssl/lib/ossl-modules"

COPY --from=build /oqs-dist /usr

COPY openssl.cnf /usr/local/ssl/openssl.cnf
ENV OPENSSL_CONF=/usr/local/ssl/openssl.cnf

WORKDIR /keygen

COPY go.mod ./
RUN go mod tidy

COPY . .
RUN CGO_ENABLED=0 go build -o keygen ./cmd/keygen

ENTRYPOINT ["./keygen"]
