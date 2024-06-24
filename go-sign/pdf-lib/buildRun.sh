#! /bin/env bash

docker build -t pdf-sign:latest -f Dockerfile.c .
docker run -v $PWD:/data -w /data pdfsign:latest\
    spec/fixtures/keys/pm-certificate.pem\
    spec/fixtures/keys/pm-private-key.pem\
    spec/fixtures/documents/new.pdf\
    spec/output/new_test_assinado.pdf