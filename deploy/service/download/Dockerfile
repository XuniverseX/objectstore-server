FROM ubuntu:20.04

ADD bin/download /

RUN chmod 777 /download

ENV PARAMS=""

ENTRYPOINT ["sh","-c","/download $PARAMS"]