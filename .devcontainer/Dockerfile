# Go version
ARG VARIANT=1
FROM mcr.microsoft.com/vscode/devcontainers/go:${VARIANT}

# Install Node.js
ARG INSTALL_NODE="true"
ARG NODE_VERSION="lts/*"
RUN if [ "${INSTALL_NODE}" = "true" ]; then su vscode -c "source /usr/local/share/nvm/nvm.sh && nvm install ${NODE_VERSION} 2>&1"; fi

# Install packages
RUN apt update \
  && apt install -y brotli

# Install a set of tools
RUN curl -sf https://gobinaries.com/github.com/ory/go-acc@v0.2.6 | PREFIX=/usr/local/bin sh \
  && curl -sf https://gobinaries.com/github.com/gobuffalo/packr/packr2@v2.7.1 | PREFIX=/usr/local/bin sh \
  && curl -sf https://gobinaries.com/github.com/markbates/pkger/cmd/pkger@v0.17.1 | PREFIX=/usr/local/bin sh \
  && curl -sf https://gobinaries.com/github.com/axw/gocov/gocov | PREFIX=/usr/local/bin sh \
  # Fork of github.com/matm/gocov-html but with tags so it works with gobinaries.com
  && curl -sf https://gobinaries.com/github.com/ItalyPaleAle/gocov-html | PREFIX=/usr/local/bin sh \
  && curl -sf https://gobinaries.com/github.com/golang/protobuf/protoc-gen-go@v1.4.3 | PREFIX=/usr/local/bin sh \
  # Protobuf compiler
  && export PROTOC_VERSION="3.14.0" \
  && PB_REL="https://github.com/protocolbuffers/protobuf/releases" \
  && curl -LO $PB_REL/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-linux-x86_64.zip \
  && unzip protoc-${PROTOC_VERSION}-linux-x86_64.zip -d protoc/ \
  && cp -rvp protoc/bin/protoc /usr/local/bin/ \
  && cp -rvp protoc/include/* /usr/local/include/ \
  && chmod 0755 /usr/local/bin/protoc \
  && chmod -R 0755 /usr/local/include/google \
  && rm -rvf protoc*

# Install global node packages: eslint
RUN su vscode -c "source /usr/local/share/nvm/nvm.sh && npm install -g eslint" 2>&1

# Configure ZSH
RUN mkdir -p /shell-history \
  && touch /shell-history/.zsh_history \
  && chown -R vscode /shell-history
COPY .zshrc /home/vscode/.zshrc

# Workdir
WORKDIR /home/vscode/workspace
