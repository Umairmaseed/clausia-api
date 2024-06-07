#if !defined(ANNOTATION_H_)
#define ANNOTATION_H_

#include <BaseDocument.hpp>
#include <VType.hpp>
#include <functional>
#include <iomanip>
#include <openssl/bio.h>
#include <openssl/cms.h>
#include <openssl/pem.h>
#include <openssl/pkcs7.h>
#include <openssl/sha.h>
#include <openssl/x509.h>
#include <podofo/podofo.h>
#include <sstream>

using namespace PoDoFo;

void print_hex(const char* string, std::string start);

enum edge { 
  first,
  last,
};

class Annotations {
private:
  /* data */
  void decodeDict(PdfDictionary& dict);
  void decodeDictPrint(PdfDictionary& dict, std::string start);
  void decodeArrayPrint(PdfArray& array, std::string start);
  void decodeRefPrint(PdfReference ref, std::string start);
  std::vector<VType> vt;
  BaseDocument* originalDoc;
  void verifyAndExtractSig(VType& val);
  std::string sha256(char* buff, int len);
  PdfVecObjects objs;

public:
  Annotations(BaseDocument* doc);
  bool Verify();
  int getSignature(char** output);
  std::vector<int> GetPendingByteRange();
  std::string GetEdgeSignatureHash(edge edge);
  ~Annotations();
  bool validateID(std::string id);
  std::string getID();
};

#endif // ANNOTATION_H_
