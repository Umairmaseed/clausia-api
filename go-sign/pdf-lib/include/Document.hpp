#ifndef __DOCUMENT__H__
#define __DOCUMENT__H__

#include <Annotations.hpp>
#include <BaseDocument.hpp>
#include <fstream>
#include <functional>
#include <openssl/pkcs7.h>
#include <pdfutils.hpp>
#include <podofo/podofo.h>
#include <vector>

using namespace std;
using namespace PoDoFo;

struct DrawInfo {
  std::string Re;
  std::string Name;
  std::string Rank;
};

struct QrCodeInfo {
  std::string url;
  std::string id;
  const unsigned char* pData;
  PoDoFo::pdf_long dwLen;
};

class Document : public BaseDocument {
private:
  bool forUpdate;
  std::vector<char> originalBuffer;

  Annotations* annots;

public:
  Document();
  void Load(string path, string password);
  void LoadFromBuffer(const char* buffer, size_t buffer_len, string password);
  bool Verify();
  void DrawQrCode(QrCodeInfo qci, DrawInfo di);
  int getSignature(char** output);
  std::string GetLastSignature();
  std::string GetFirstSignature();
  std::vector<int> PendingByteRange();
  bool IsValidID(std::string id);
  std::string getID();
  ~Document();
};
#endif //!__DOCUMENT__H__
