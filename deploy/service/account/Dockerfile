FROM ubuntu:20.04

ADD bin/account /

RUN chmod 777 /account

ENV PARAMS=""

ENTRYPOINT ["sh","-c","/account $PARAMS"]