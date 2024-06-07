# pdf-sign

## Prerequisites

- Podofo lib
- OpenSSL
- Catch2 (for unit tests) (tag v3.0.0-preview3)

## Configuration

### Final Cert

Final cert is the pair used for locking down files from new signatures after that
You can configure final cert pair in certain ways:

If you're invoking this code through a shared object/dynamic library you can overwrite following variables

```bash 
export FINAL_PUBLIC_KEY=/path/to/public/key
export FINAL_PRIVATE_KEY=/path/to/private/key
export FINAL_PASSWORD=PasswordValue
```

If you're developing using this code as a package, there are some constructors for Signer that you can use.

Finally if both environment variables and properly constructor isn't used, then it will fall to default paths `config/final_key/public.pem` `config/final_key/private.pem`


## Using docker

```
docker build -t pdf-sign:latest -f Dockerfile.c .
docker run -v $PWD:/data -w /data pdfsign:latest spec/fixtures/keys/pm-certificate.pem  spec/fixtures/keys/pm-private-key.pem  spec/fixtures/documents/new.pdf spec/fixtures/output/new_test_assinado.pdf
```

```
cmake --build build
sudo cmake --build build --target install -- -j $(nproc)
```
