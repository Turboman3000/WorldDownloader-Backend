FROM debian

COPY ./ /srv

CMD ["/srv/wdl-backend"]

EXPOSE 8080