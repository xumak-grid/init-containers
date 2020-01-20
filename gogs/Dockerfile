FROM registry.xumak.gt:5000/xumak/maven:3.3.9-jdk8
LABEL maintainer = "ehernandez@xumak.com"
RUN apk --no-cache add openssh-client git unzip

ENV GOGS_HOST=localhost
ENV GOGS_CONFIG_FILE=/app/examples/configFile.json

# Disable stric host check
COPY ssh_config /root/.ssh/config
# Add deploy key to xumak-grid/demo
COPY id_rsa /root/.ssh/id_rsa
RUN chmod 600 /root/.ssh/id_rsa

WORKDIR /app
COPY examples /app/examples/
COPY resources/danta-aem-archetype /root/.m2/repository/io/tikaltechnologies/danta/danta-aem-archetype
COPY bin/init-gogs /app/init-gogs
CMD [ "/app/init-gogs" ]
