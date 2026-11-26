FROM alpine:latest
MAINTAINER zhufuyi "g.zhufuyi@gmail.com"

# add curl, used for http service checking, can be installed without it if deployed in k8s
RUN apk add curl

COPY configs/ /app/configs/
COPY casino /app/casino
RUN chmod +x /app/casino

# http port
EXPOSE 8080


WORKDIR /app

CMD ["./casino", "-c", "configs/casino.yml"]
# if you use the Configuration Center, casino.yml is changed to the Configuration Center configuration.
#CMD ["./casino", "-c", "configs/casino.yml", "-enable-cc"]
