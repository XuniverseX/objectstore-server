FROM ubuntu:20.04

ADD bin/transfer /

RUN chmod 777 /transfer

ENV PARAMS=""

ENTRYPOINT ["sh","-c","/transfer $PARAMS"]