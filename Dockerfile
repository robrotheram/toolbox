FROM index.docker.io/codercom/enterprise-base:ubuntu
LABEL org.opencontainers.image.source="https://github.com/robrotheram/toolbox"
ENV GO_VERSION="1.21.4"
ENV GO_ARCH="amd64"
ENV HELM_VERSION="3.12.0"
ENV KUBE_VERSION="v1.27.1"
ENV NODE_VERSION="18.3.0"
ENV HUGO_VERSION="0.111.3"
ENV PATH="${PATH}:/usr/local/go/bin:/usr/local/nodejs/bin"

USER root
RUN apt-get update && apt-get install unzip zip jq -y && rm -rf /var/lib/apt/lists/*

RUN curl -O "https://nodejs.org/dist/v${NODE_VERSION}/node-v${NODE_VERSION}-linux-x64.tar.xz" \
    && tar -xf "node-v${NODE_VERSION}-linux-x64.tar.xz" \
    && mv "node-v${NODE_VERSION}-linux-x64" /usr/local/nodejs \
    && rm "node-v${NODE_VERSION}-linux-x64.tar.xz"
RUN curl -O -L "https://github.com/gohugoio/hugo/releases/download/v${HUGO_VERSION}/hugo_extended_${HUGO_VERSION}_linux-amd64.deb" && dpkg -i hugo_extended_${HUGO_VERSION}_linux-amd64.deb && rm hugo_extended_${HUGO_VERSION}_linux-amd64.deb
RUN curl -O -L "https://golang.org/dl/go${GO_VERSION}.linux-${GO_ARCH}.tar.gz" && tar -C /usr/local -xzf "go${GO_VERSION}.linux-${GO_ARCH}.tar.gz" && rm "go${GO_VERSION}.linux-${GO_ARCH}.tar.gz"
RUN curl -s "https://get.sdkman.io" | bash
RUN curl -LO "https://dl.k8s.io/release/${KUBE_VERSION}/bin/linux/amd64/kubectl" && chmod +x ./kubectl && mv ./kubectl /usr/local/bin
RUN curl -O -L "https://get.helm.sh/helm-v${HELM_VERSION}-linux-amd64.tar.gz" && tar zxvf "helm-v${HELM_VERSION}-linux-amd64.tar.gz" && mv linux-amd64/helm  /usr/local/bin/helm && chmod +x /usr/local/bin/helm && rm -rf linux-amd64 && rm "helm-v${HELM_VERSION}-linux-amd64.tar.gz"

# Set back to coder user
USER coder


