FROM alpine:3.7
LABEL maintainer = "ehernandez@xumak.com"
RUN apk --no-cache add ca-certificates

ENV NEXUS_USER=admin
ENV NEXUS_PASS=admin123
ENV NEXUS_HOST=localhost
ENV NEXUS_CONFIG_FILE=/app/examples/configFile.json

WORKDIR /app
COPY examples/configFile.json /app/examples/configFile.json
COPY bin/init-nexus /app/init-nexus
CMD [ "/app/init-nexus" ]
