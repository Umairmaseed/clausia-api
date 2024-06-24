FROM alpine:3.12

RUN apk add g++ make cmake openssl-dev podofo-dev git catch2
RUN mkdir -p /build
ADD . /build/signer

WORKDIR /build/signer

RUN mkdir -p build
RUN rm -rf build/*

RUN sh -c 'cd build && cmake .. && make -j2'
RUN cp build/pdf-sign /usr/local/bin/pdf-sign

RUN cd build && ./pdfSign_test

ENTRYPOINT [ "pdf-sign" ]