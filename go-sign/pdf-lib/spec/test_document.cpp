#include "Document.hpp"
#include <Annotations.hpp>
#include <Signer.hpp>
#include <catch2/catch.hpp>
#include <fstream>
#include <podofo/podofo.h>
#include <string>
#include <testUtils.hpp>
#include <test_const.h>

TEST_CASE("Load file from disk") {
  auto doc = new Document();

  REQUIRE_NOTHROW(doc->Load(documents_path + "new.pdf", ""));

  // load for update
  REQUIRE_NOTHROW(doc->Load(documents_path + "new.pdf", ""));
}

TEST_CASE("Load encrypted for update") {
  auto doc = new Document();
  try {
    doc->Load(documents_path + "blank-password.pdf", "pwd");
  } catch (PoDoFo::PdfError& e) {
    REQUIRE(e.GetError() == PoDoFo::ePdfError_CannotEncryptedForUpdate);
  }
}

TEST_CASE("Verify signed pdf") {
  auto signedDocBinary = CreateDocSigned("blank.pdf");

  auto newDoc = Document();
  newDoc.LoadFromBuffer(signedDocBinary.c_str(), signedDocBinary.length(), "");
  newDoc.Verify();

  REQUIRE_NOTHROW(newDoc.Verify());
  REQUIRE(newDoc.Verify());
}

TEST_CASE("Get Signature") {
  auto signedDocBinary = CreateDocSigned("blank.pdf");

  auto newDoc = Document();
  newDoc.LoadFromBuffer(signedDocBinary.c_str(), signedDocBinary.length(), "");

  char* sigHash = {};
  sigHash = (char*)malloc(30000);
  int ret = newDoc.getSignature(&sigHash);

  REQUIRE(ret > 0);
}

TEST_CASE("Test Edge signatures") {
  auto doc = CreateNewSignedDocObj("blank.pdf");

  auto lastWritten = doc->GetLastSignature();

  auto firstSig = doc->GetFirstSignature();

  REQUIRE_FALSE(lastWritten.empty());
  REQUIRE_FALSE(firstSig.empty());
  REQUIRE(lastWritten == firstSig);
}

TEST_CASE("Test Final with 1 sig") {
  std::cout << "final true test" << std::endl;
  auto doc = CreateNewSignedDocObj("blank_final.pdf");

  auto lastWritten = doc->GetLastSignature();

  auto firstSig = doc->GetFirstSignature();

  std::cout << lastWritten << std::endl;

  REQUIRE_FALSE(lastWritten.empty());
  REQUIRE_FALSE(firstSig.empty());
  REQUIRE_FALSE(lastWritten == firstSig);
}

TEST_CASE("Test first sig for double-signed docs") {
  auto firstDoc = new Document();

  firstDoc->Load(documents_path + "blank_signed.pdf", "");

  auto sndDoc = new Document();
  sndDoc->Load(documents_path + "blank_signed_.pdf", "");

  auto signature = firstDoc->GetFirstSignature();
  REQUIRE_FALSE(signature.empty());
  auto newSig = sndDoc->GetFirstSignature();

  REQUIRE_FALSE(newSig.empty());

  REQUIRE(newSig == signature);
}

TEST_CASE("Test first is different from snd") {
  auto doc = new Document();
  doc->Load(documents_path + "blank_signed_.pdf", "");

  auto signature = doc->GetFirstSignature();
  REQUIRE_FALSE(signature.empty());
  auto newSig = doc->GetLastSignature();

  REQUIRE_FALSE(newSig.empty());

  REQUIRE_FALSE(newSig == signature);
}

TEST_CASE("Test document qrcode draw") {
  auto doc = new Document();
  doc->Load(documents_path + "pm example.pdf", "");

  std::ifstream qrcodeFile(qrCodePath, ios_base::binary);

  FILE* f = fopen(qrCodePath.c_str(), "rb");
  fseek(f, 0, SEEK_END);
  long fsize = ftell(f);
  fseek(f, 0, SEEK_SET); /* same as rewind(f); */

  unsigned char* qrCode = (unsigned char*)malloc(fsize + 1);
  fread(qrCode, 1, fsize, f);
  fclose(f);

  REQUIRE(fsize > 0);

  auto qci = QrCodeInfo{
    url : "https://policiamilitar.example.sp.gov.br/",
    id : "example-78e7-4a5b-a24d-32b2f699e354",
    pData : qrCode,
    dwLen : fsize,
  };
  auto di = DrawInfo{
    Re : "1",
    Name : "Fulano Teste",
    Rank : "Tenente",
  };

  doc->DrawQrCode(qci, di);

  auto buf = doc->GetDocBuffer();
  REQUIRE_FALSE(buf.empty());

  std::ofstream outDoc(output_path + "qrCodeDraw.pdf");
  outDoc << std::string(buf.begin(), buf.end());
  outDoc.close();
}