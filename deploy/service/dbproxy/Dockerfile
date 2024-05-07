FROM ubuntu:20.04

ADD bin/dbproxy /

RUN chmod 777 /dbproxy

ENV PARAMS=""

ENTRYPOINT ["sh","-c","/dbproxy $PARAMS"]