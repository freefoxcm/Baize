# syntax=docker/dockerfile:1.7

ARG GO_IMAGE=golang:1.26.5-bookworm
ARG NODE_IMAGE=node:22-bookworm-slim

FROM ${GO_IMAGE} AS builder

ARG BUILD_VERSION=dev
ARG BUILD_COMMIT=unknown
ARG BUILD_TIME_UTC=unknown
ARG GOPROXY=https://goproxy.cn,direct
ENV CGO_ENABLED=0 GOPROXY=${GOPROXY}

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -trimpath \
    -ldflags "-s -w -X main.version=${BUILD_VERSION} -X main.gitCommit=${BUILD_COMMIT} -X main.buildTimeUTC=${BUILD_TIME_UTC}" \
    -o /out/reasonix ./cmd/reasonix

FROM ${NODE_IMAGE} AS runtime

ARG DEBIAN_MIRROR=https://mirrors.aliyun.com/debian
ARG DEBIAN_SECURITY_MIRROR=https://mirrors.aliyun.com/debian-security
ARG PYPI_INDEX=https://pypi.tuna.tsinghua.edu.cn/simple
ARG NPM_REGISTRY=https://registry.npmmirror.com
ARG NODE_DIST_URL=https://npmmirror.com/mirrors/node
ARG PNPM_VERSION=10.34.5
ARG APP_UID=10001
ARG APP_GID=10001

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

RUN rm -f /etc/apt/sources.list.d/debian.sources \
    && printf '%s\n' \
      "deb ${DEBIAN_MIRROR} bookworm main contrib non-free non-free-firmware" \
      "deb ${DEBIAN_MIRROR} bookworm-updates main contrib non-free non-free-firmware" \
      "deb ${DEBIAN_SECURITY_MIRROR} bookworm-security main contrib non-free non-free-firmware" \
      > /etc/apt/sources.list \
    && apt-get update \
    && apt-get install -y --no-install-recommends \
      bash bubblewrap build-essential ca-certificates chromium curl fonts-noto-cjk git jq make \
      openssh-client pkg-config python-is-python3 python3 python3-pip python3-venv \
      ripgrep tar unzip wget zip \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /usr/local/go /usr/local/go
COPY --from=builder /out/reasonix /usr/local/bin/reasonix
COPY docker/requirements-skills.lock /tmp/requirements-skills.lock
COPY docker/reasonix-entrypoint.sh /usr/local/bin/reasonix-entrypoint

RUN printf '[global]\nindex-url = %s\ntimeout = 60\nretries = 5\n' "${PYPI_INDEX}" > /etc/pip.conf \
    && printf 'registry=%s\ndisturl=%s\nfetch-retries=5\nfetch-timeout=60000\n' "${NPM_REGISTRY}" "${NODE_DIST_URL}" > /etc/npmrc \
    && python3 -m venv /opt/reasonix-skills \
    && /opt/reasonix-skills/bin/pip install --no-cache-dir --require-hashes -r /tmp/requirements-skills.lock \
    && npm install --global --registry "${NPM_REGISTRY}" "pnpm@${PNPM_VERSION}" \
    && groupadd --gid "${APP_GID}" reasonix \
    && useradd --uid "${APP_UID}" --gid "${APP_GID}" --create-home reasonix \
    && mkdir -p /workspace /var/lib/reasonix /opt/reasonix-runtime \
    && chown -R reasonix:reasonix /workspace /var/lib/reasonix /opt/reasonix-runtime \
    && chmod 0755 /usr/local/bin/reasonix-entrypoint \
    && rm -f /tmp/requirements-skills.lock

ENV REASONIX_HOME=/var/lib/reasonix \
    GOPROXY=https://goproxy.cn,direct \
    GOPATH=/opt/reasonix-runtime/go \
    GOMODCACHE=/opt/reasonix-runtime/go/pkg/mod \
    GOCACHE=/opt/reasonix-runtime/go-cache \
    PIP_CONFIG_FILE=/etc/pip.conf \
    PIP_TARGET=/opt/reasonix-runtime/python-site \
    PYTHONPATH=/opt/reasonix-runtime/python-site \
    NPM_CONFIG_USERCONFIG=/etc/npmrc \
    NPM_CONFIG_PREFIX=/opt/reasonix-runtime/npm \
    NPM_CONFIG_CACHE=/opt/reasonix-runtime/npm-cache \
    PNPM_HOME=/opt/reasonix-runtime/pnpm \
    PNPM_STORE_DIR=/opt/reasonix-runtime/pnpm-store \
    PATH=/opt/reasonix-runtime/npm/bin:/opt/reasonix-runtime/pnpm:/opt/reasonix-skills/bin:/usr/local/go/bin:${PATH}

WORKDIR /workspace
USER reasonix
EXPOSE 8787
ENTRYPOINT ["reasonix-entrypoint"]
CMD ["serve", "--addr", "0.0.0.0:8787", "--dir", "/workspace", "--no-open"]
