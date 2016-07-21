FROM scratch
MAINTAINER Steve Sloka <steve@stevesloka.com>
ADD certs/ certs/
ADD restapi restapi
ENTRYPOINT ["/restapi"]
