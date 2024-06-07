#if !defined(VTYPE_H_)
#define VTYPE_H_

#include <openssl/pkcs7.h>
#include <openssl/x509.h>
#include <string>
#include <vector>

struct VType {
  struct {
    PKCS7* p7;
    unsigned char* messageHash;
    int messageLength;
    STACK_OF(X509) * certs;
    std::string Location;
    std::shared_ptr<ASN1_UTCTIME*> timestamp;
  } Signature;
  std::vector<int> ByteRange;
  char* Contents;
  int contentsLen;
  std::string Location;
  struct {
    struct {
      std::string Name;
    } App;
  } Prop_Build;
  std::string Reason;
  std::string SubFilter;
  std::string Type;
  std::string DocumentID;
  bool pending = false;
};

#endif // VTYPE_H_
