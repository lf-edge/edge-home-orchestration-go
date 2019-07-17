# Docker image for "edge-orchestration"
# TODO - need to reduce base image of edge-orchestration
### ubuntu:16.04 image size is 119MB
### alpine:3.6 image size is 4MB
FROM ubuntu:16.04

# environment variables
ENV TARGET_DIR=/edge-orchestration
ENV HTTP_PORT=56001
ENV MDNS_PORT=5353
ENV ZEROCONF_PORT=42425
ENV APP_MAIN_DIR=GoMain
ENV APP_BIN_DIR=$APP_MAIN_DIR/bin
ENV APP_NAME=edge-orchestration

# copy files
COPY $APP_BIN_DIR/$APP_NAME $APP_MAIN_DIR/run.sh $TARGET_DIR/

# expose ports
EXPOSE $HTTP_PORT $MDNS_PORT $ZEROCONF_PORT

# set the working directory
WORKDIR $TARGET_DIR

# make boltdb volume
RUN mkdir -p /var/data/db

# kick off the edge-orchestration container
CMD ["sh", "run.sh"]
