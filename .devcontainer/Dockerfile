# Go version
ARG VARIANT=1
FROM mcr.microsoft.com/vscode/devcontainers/go:${VARIANT}

# Install Node.js
ARG INSTALL_NODE="true"
ARG NODE_VERSION="12"
RUN if [ "${INSTALL_NODE}" = "true" ]; then su vscode -c "source /usr/local/share/nvm/nvm.sh && nvm install ${NODE_VERSION} 2>&1"; fi

# Install a set of tools
RUN curl -sf https://gobinaries.com/github.com/ory/go-acc@v0.2.6 | PREFIX=/usr/local/bin sh \
  && curl -sf https://gobinaries.com/github.com/ory/go-acc@v0.2.6 | PREFIX=/usr/local/bin sh \
  && curl -sf https://gobinaries.com/github.com/gobuffalo/packr/packr2@v2.7.1 | PREFIX=/usr/local/bin sh \
  && curl -sf https://gobinaries.com/github.com/markbates/pkger/cmd/pkger@v0.17.1 | PREFIX=/usr/local/bin sh

# Install global node packages: eslint
RUN su vscode -c "source /usr/local/share/nvm/nvm.sh && npm install -g eslint" 2>&1
