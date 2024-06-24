FROM alpine:3.13 as builder

RUN apk add g++ make cmake openssl-dev podofo-dev git catch2 sudo bash curl gcc

COPY pdf-lib /pdf-lib/
WORKDIR /pdf-lib
RUN rm -rf build
RUN mkdir build
RUN cd build && cmake ..
RUN cmake --build build --target pdfsign install -- -j $(nproc)

FROM golang:1.17-alpine3.13

RUN apk add gcc musl-dev openssl-dev podofo-dev curl fontconfig

COPY --from=builder /usr/local/lib64/libpdfsign.so.1.0 /usr/local/lib64/libpdfsign.so.1.0
COPY --from=builder /usr/local/lib64/libpdfsign.so.1 /usr/local/lib64/libpdfsign.so.1
COPY --from=builder /usr/local/lib64/libpdfsign.so /usr/local/lib64/libpdfsign.so
COPY --from=builder /usr/local/include/pdfsign.h /usr/local/include/pdfsign.h
COPY --from=builder /usr/local/include/wrapper.hpp /usr/local/include/wrapper.hpp
COPY --from=builder /usr/local/share/pkgconfig/pdfsign.pc /usr/local/share/pkgconfig/pdfsign.pc

ENV LD_LIBRARY_PATH=/usr/local/lib64/:${LD_LIBRARY_PATH}
ENV GIN_MODE=release

RUN mkdir -p /usr/share/fonts
COPY ./fonts/Roboto-Light.ttf /usr/share/fonts
COPY ./fonts/Roboto-Thin.ttf /usr/share/fonts
COPY ./fonts/OpenSans-Light.ttf /usr/share/fonts

RUN fc-cache -f -v

COPY . /go-sign
WORKDIR /go-sign
RUN go mod vendor
RUN go build -mod=vendor
CMD [ "./go-sign" ]