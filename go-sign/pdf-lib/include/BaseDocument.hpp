#ifndef __BASEDOCUMENT__H__
#define __BASEDOCUMENT__H__

#include <iomanip>
#include <math.h>
#include <openssl/bio.h>
#include <podofo/podofo.h>
#include <regex>
#include <sstream>
#include <string>

#include <vector>

enum DocumentInputDevice { Memory };

using std::string;
using namespace PoDoFo;

class BaseDocument {
public:
  BaseDocument(std::vector<char>& originalRef);
  ~BaseDocument();

  PoDoFo::PdfDocument* Base;
  PoDoFo::PdfRefCountedBuffer* StreamDocRefCountedBuffer = nullptr;
  PoDoFo::PdfOutputDevice* StreamDocOutputDevice = nullptr;
  std::vector<char>& originalRef;
  inline PoDoFo::PdfMemDocument& GetDocument() const {
    return *dynamic_cast<PoDoFo::PdfMemDocument*>(Base);
  }

  PoDoFo::PdfRefCountedInputDevice& input;

  int GetDocumentSliced(std::vector<int> byteRange, BIO* out);
  std::vector<char> GetDocBuffer();
  std::vector<int> GetSignData(BIO* dataBuffer);

protected:
  void load(string filePath,
            PdfRefCountedInputDevice* input,
            string pwd,
            DocumentInputDevice deviceType);
  std::vector<PoDoFo::PdfObject*> Copies;
};

#endif //!__BASEDOCUMENT__H__
