FROM ubuntu:latest
LABEL authors="dfrishchin"

ENTRYPOINT ["top", "-b"]