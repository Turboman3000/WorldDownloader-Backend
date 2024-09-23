FROM debian

COPY ./ /srv

RUN apt update && apt install curl -y

CMD ["/srv/wdl-backend"]

EXPOSE 8080