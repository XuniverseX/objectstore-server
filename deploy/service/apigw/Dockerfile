FROM ubuntu:20.04

ADD bin/apigw /

RUN chmod 777 /apigw

ENV PARAMS=""

ENTRYPOINT ["sh","-c","/apigw $PARAMS"]