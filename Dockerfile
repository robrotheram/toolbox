FROM index.docker.io/codercom/enterprise-base:ubuntu
# Run everything as root
USER root

LABEL org.opencontainers.image.source="https://github.com/robrotheram/toolbox"
RUN apt-get update && apt-get install unzip zip jq -y

ENV GO_VERSION="1.20.4"
ENV GO_ARCH="amd64"
ENV HELM_VERSION="3.12.0"
ENV KUBE_VERSION="v1.27.1"
ENV NVM_VERSION="0.35.3"
ENV PATH="${PATH}:/usr/local/go/bin"


RUN curl -O -L "https://golang.org/dl/go${GO_VERSION}.linux-${GO_ARCH}.tar.gz" && tar -C /usr/local -xzf "go${GO_VERSION}.linux-${GO_ARCH}.tar.gz"
RUN curl -s "https://get.sdkman.io" | bash
RUN curl -o- "https://raw.githubusercontent.com/nvm-sh/nvm/v${NVM_VERSION}/install.sh" | bash
RUN curl -LO "https://dl.k8s.io/release/${KUBE_VERSION}/bin/linux/amd64/kubectl" && chmod +x ./kubectl && mv ./kubectl /usr/local/bin
RUN curl -O -L "https://get.helm.sh/helm-v${HELM_VERSION}-linux-amd64.tar.gz" && tar zxvf "helm-v${HELM_VERSION}-linux-amd64.tar.gz" && mv linux-amd64/helm  /usr/local/bin/helm && chmod +x /usr/local/bin/helm && rm -rf linux-amd64

# Set back to coder user
USER coder

