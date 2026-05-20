FROM golang:1.26-alpine AS build-app

ARG APP_NAMESPACE
ARG APP_NAME
ARG GOPRIVATE
ENV APP_NAMESPACE=$APP_NAMESPACE
ENV APP_NAME=$APP_NAME
ENV GOPRIVATE=$GOPRIVATE

WORKDIR /opt/${APP_NAMESPACE}
RUN apk add --no-cache git make
COPY ./ ./
RUN --mount=type=secret,id=NETRC,target=/root/.netrc \
    make build


FROM alpine:3.21 AS pack-image

ARG APP_NAMESPACE
ARG APP_NAME
ENV APP_NAMESPACE=$APP_NAMESPACE
ENV APP_NAME=$APP_NAME
ENV WORK_DIR=/opt/${APP_NAMESPACE}

WORKDIR ${WORK_DIR}
RUN apk add --no-cache gcompat libc6-compat && \
    ln -s /lib/libc.so.6 /usr/lib/libresolv.so.2

COPY --from=build-app ${WORK_DIR}/.dist/${APP_NAME} .
RUN chmod u+x ./${APP_NAME}

ENTRYPOINT ["sh", "-c", "${WORK_DIR}/${APP_NAME}"]
