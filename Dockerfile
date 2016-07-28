FROM scratch
MAINTAINER Steve Sloka <steve@stevesloka.com>
ADD certs/ certs/
ADD db db/
ADD restapi restapi
ENTRYPOINT ["/restapi"]
