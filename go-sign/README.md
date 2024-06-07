# go-sign

go-sign is a document signer API that uses x.509 certificates. Its meant to be used either by general purposes or alongside a chaincode.

## Setting up

After cloning the repository and getting the Docker Image, use the compose file for setting up the API.

```bash
docker-compose up --build  # first run
docker-compose up          # after building in first run
```

This will download all dependencies and let the API good to go on port `:8080`.

## Endpoints

The API has three main operations: salting, signing and verifying PDF file.

### Salt PDF

For salting a PDF file, you may use the `/api/saltPdf` route. The request is  a `multipart/file-form` with the following fields:

- `file`: the PDF file to be salted;

Example `curl` request:

```bash
curl -v -F file=@/path/to/file.pdf http://localhost:8080/api/saltPdf
```

### Sign PDF

For signing a PDF file, you may use the `/api/signdocs` route. The request is  a `multipart/file-form` with the following fields:

- `file`: the PDF file to be signed;
- `fileName`: the PDF file name;
- `certificate`: the certificate file (e.g pfx, pem);
- `password`: certificate's password;
- `ledgerKey`: this is a supplementary field to be used in case of the document is being stored in a blockchain network. This is its unique id in ledger;
- `clientBaseUrl`: this is the url used for generating QR code;
- `signature`: an object describing digital signature's position. It can be passed as string and has the following structure:

```json
{ 
    "file.pdf": { 
        "rect": { 
            "x": 309.609375, 
            "y": 403.33333333333337, 
            "page": 1 
        }, 
    "final": false 
    } 
}
```

Example `curl` request:

```bash
curl -v \
> -F file=@/path/to/file.pdf \
> -F fileName='file.pdf' \
> -F certificate=@/path/to/certificate.pfx \
> -F password='cert_password' \
> -F ledgerKey='document:xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx' \
> -F signature='{ "file.pdf": { "rect": { "x": 309.609375, "y": 403.33333333333337, "page": 1 }, "final": false } }' \
> -F clientBaseUrl='localhost/verify' \
> http://localhost:8080/api/signdocs

```

### Verify PDF

For verifying a signed PDF file, you may use the `/api/verifydocs` route. The request is  a `multipart/file-form` with the following fields:

- `files`: the PDF file(s) to be verified;

Example `curl` request:

```bash
curl -v -F files=@/path/to/file.pdf http://localhost:8080/api/verifydocs
```

### Get Key

For getting a ledger key attached to a PDF (if it has), you may use the `/api/getkey` route. The request is  a `multipart/file-form` with the following fields:

- `file`: the PDF file from which the ledger key will extracted;

Example `curl` request:

```bash
curl -v -F file=@/path/to/file.pdf http://localhost:8080/api/getkey
```

## Testing

Unit test files are located at `api` folder. It's highly recommended that `.devcontainer` is used for testing, for compilation issues to be avoided.

Inside `.devcontainer` you can use the following command to execute all tests:

```bash
go test ./...
```
