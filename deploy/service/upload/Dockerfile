FROM ubuntu:20.04

ADD bin/upload /

RUN chmod 777 /upload

ENV PARAMS=""

ENTRYPOINT ["sh","-c","/upload $PARAMS"]