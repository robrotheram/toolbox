FROM index.docker.io/codercom/enterprise-base:ubuntu
# Run everything as root
USER root

LABEL org.opencontainers.image.source="https://github.com/robrotheram/toolbox"
RUN apt-get update && apt-get install unzip zip jq -y
ENV VERSION="1.20.4"
ENV ARCH="amd64"

RUN curl -O -L "https://golang.org/dl/go${VERSION}.linux-${ARCH}.tar.gz" && tar -C /usr/local -xzf "go${VERSION}.linux-${ARCH}.tar.gz"
RUN curl -s "https://get.sdkman.io" | bash
RUN curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.35.3/install.sh | bash
RUN curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" && chmod +x ./kubectl && mv ./kubectl /usr/local/bin
RUN curl -O -L "https://get.helm.sh/helm-v3.12.0-linux-amd64.tar.gz" && tar zxvf "helm-v3.12.0-linux-amd64.tar.gz" && mv linux-amd64/helm  /usr/local/bin/helm && chmod +x /usr/local/bin/helm && rm -rf linux-amd64
ENV PATH="${PATH}:/usr/local/go/bin"

# Set back to coder user
USER coder

