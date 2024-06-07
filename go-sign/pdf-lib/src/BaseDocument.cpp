#include "BaseDocument.hpp"
#include <iostream>

using namespace PoDoFo;

BaseDocument::BaseDocument(std::vector<char>& original)
  : originalRef(original)
  , input(*(new PdfRefCountedInputDevice())) {
  input = PdfRefCountedInputDevice();
  Base = new PdfMemDocument();
}

BaseDocument::~BaseDocument() {
  for (auto c : Copies) {
    delete c;
  }
  delete Base;
  delete StreamDocOutputDevice;
  delete StreamDocRefCountedBuffer;
}

void BaseDocument::load(string filePath,
                        PdfRefCountedInputDevice* input,
                        string pwd,
                        DocumentInputDevice deviceType) {
  auto& doc = GetDocument();
  try {
    switch (deviceType) {
      case DocumentInputDevice::Memory:
        doc.LoadFromDevice(*input, true);
        this->input = *input;
    }
  } catch (PdfError& e) {
    if (e.GetError() == ePdfError_InvalidPassword) {
      if (pwd.empty()) {
        throw std::runtime_error("Password required for this doc");
      } else {
        try {
          doc.SetPassword(pwd);
        } catch (PdfError& e) {
          throw e;
        }
      }
    } else {
      std::cout << e.ErrorMessage(e.GetError()) << std::endl;
      throw e;
    }
  }
}

int BaseDocument::GetDocumentSliced(std::vector<int> byteRange, BIO* out) {
  if (originalRef.size() > 0 && byteRange.size() == 4) {
    int readLen = byteRange[1] - byteRange[0];
    auto size = originalRef.size();
    auto readBuf = std::string(originalRef.begin() + byteRange[0],
                               originalRef.begin() + byteRange[1]);

    int written = BIO_write(out, readBuf.c_str(), readBuf.size());

    readLen = byteRange[3] - byteRange[2];

    readBuf = std::string(originalRef.begin() + byteRange[2],
                          originalRef.begin() + byteRange[2] + byteRange[3]);

    written += BIO_write(out, readBuf.c_str(), readBuf.size());

    int lastByte = byteRange[2] + byteRange[3];
    if (written > lastByte) {
      throw std::runtime_error("Invalid document slicing: Out of bounds.");
    }

    return written;
  }
  return 0;
}

std::vector<int> BaseDocument::GetSignData(BIO* dataBuffer) {
  auto buf = GetDocBuffer();
  std::string pdfStr(buf.begin(), buf.end() - 1);

  const auto placeholder =
    std::string("/ByteRange[ 0 1234567890 1234567890 1234567890]");
  auto brPos = pdfStr.find(placeholder);
  auto brEnd = brPos + 32;

  auto contentsTagPos = pdfStr.find("/Contents", brEnd);
  auto placeholderPos = pdfStr.find("<", contentsTagPos);
  auto placeholderEnd = pdfStr.find(">", placeholderPos);

  auto placeholderLengthWithBrackets = (placeholderEnd + 1) - placeholderPos;
  auto placeholderLength = placeholderLengthWithBrackets - 2;

  std::vector<int> byteRange = { 0, 0, 0, 0 };
  byteRange[1] = placeholderPos;
  byteRange[2] = byteRange[1] + placeholderLengthWithBrackets;
  byteRange[3] = (pdfStr.length()) - byteRange[2];

  char buffer[99];
  sprintf(buffer,
          "/ByteRange[ %d %d %d %d",
          byteRange[0],
          byteRange[1],
          byteRange[2],
          byteRange[3]);
  std::string actualByteRange = buffer;
  actualByteRange +=
    std::string(abs(int(placeholder.length() - actualByteRange.length())),
                ' ') +
    ']';

  pdfStr.replace(brPos, actualByteRange.length(), actualByteRange);
  auto toSign = pdfStr.substr(0, byteRange[1]) +
                pdfStr.substr(byteRange[2], byteRange[2] + byteRange[3]);

  BIO_write(dataBuffer, toSign.c_str(), toSign.size());

  return byteRange;
}

std::vector<char> BaseDocument::GetDocBuffer() {
  PdfRefCountedBuffer Buffer;
  PdfOutputDevice* outDev = new PdfOutputDevice(&Buffer);

  outDev->Seek(0);
  GetDocument().Write(outDev);

  return std::vector<char>(Buffer.GetBuffer(),
                           Buffer.GetBuffer() + Buffer.GetSize());
}