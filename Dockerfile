FROM golang:1.20.5-bullseye AS BUILD

RUN apt-get update && apt-get install -y make  && apt-get install -y gcc

COPY . /simple-csi-driver

WORKDIR /simple-csi-driver

RUN make build

RUN chmod a+x /simple-csi-driver/simple-csi-driver

FROM registry.k8s.io/build-image/debian-base:bullseye-v1.4.3

COPY --from=BUILD /simple-csi-driver/simple-csi-driver /simple-csi-driver

LABEL chenliu1993="cl2037829916@gmail.com"

ENTRYPOINT ["/simple-csi-driver"]

