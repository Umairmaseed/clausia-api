#ifndef __SIGNER__H__
#define __SIGNER__H__

#include <functional>
#include <string>

#include <podofo/podofo.h>

#include "names.hpp"
#include <Annotations.hpp>
#include <Document.hpp>
#include <cstdlib>
#include <ctime>
#include <iostream>
#include <openssl/crypto.h>
#include <openssl/err.h>
#include <openssl/evp.h>
#include <openssl/pem.h>
#include <openssl/pkcs12.h>
#include <openssl/pkcs7.h>
#include <openssl/ssl.h>
#include <openssl/x509.h>
#include <podofo/podofo.h>

class Signer {
private:
  std::string Output;

  PoDoFo::pdf_int32 MinSigSize = 0;
  EVP_PKEY* Pkey = nullptr;
  X509* Cert = nullptr;

  EVP_PKEY* FinalPkey = nullptr;
  X509* FinalCert = nullptr;

  PoDoFo::pdf_int32 minSignerSize(FILE* fp);
  void draw(std::string path);
  void loadFinalPair(std::string, std::string, std::string);
  Document& Doc;

public:
  Signer(Document& document, std::string output);
  void LoadPairFromMemory(std::string pub,
                          std::string priv,
                          std::string password);
  void LoadPairFromDisk(std::string pub,
                        std::string priv,
                        std::string password);
  bool isPairLoaded();

  int sign(char** write_buffer, std::string path);

  std::shared_ptr<PoDoFo::PdfSignatureField> SignatureField;
  ~Signer();
};

#endif //!__SIGNER__H__
