#include <testUtils.hpp>

std::shared_ptr<PoDoFo::PdfSignatureField> createField(Document* doc) {
  auto rect = PoDoFo::PdfRect(10, 100, 200, 100);
  auto nPages = doc->GetDocument().GetPageCount();
  auto page = doc->GetDocument().GetPage(nPages - 1);
  auto annot =
    page->CreateAnnotation(PoDoFo::EPdfAnnotation::ePdfAnnotation_Widget, rect);

  annot->SetFlags(PoDoFo::EPdfAnnotationFlags::ePdfAnnotationFlags_Print);

  auto field =
    std::shared_ptr<PoDoFo::PdfSignatureField>(new PoDoFo::PdfSignatureField(
      annot, doc->GetDocument().GetAcroForm(), &doc->GetDocument()));
  field->SetSignatureReason(PoDoFo::PdfString("razao"));
  field->SetSignatureLocation(PoDoFo::PdfString("here"));
  field->SetSignatureCreator(PoDoFo::PdfName("pmesp.signer"));
  field->SetFieldName("pmesp.signer");
  field->SetSignatureDate(PoDoFo::PdfDate());
  field->SetAlternateName("alternate_id");

  return field;
}

std::string CreateDocSigned(std::string doc_path) {
  auto doc = new Document();

  doc->Load(documents_path + doc_path, "");

  auto signer = Signer(*doc, "");
  auto field = createField(doc);
  signer.SignatureField = field;

  REQUIRE_NOTHROW(signer.LoadPairFromDisk(
    keys_path + "certificate.pem", keys_path + "private-key.pem", ""));

  char* output = {};
  output = (char*)malloc(650000);
  auto written = signer.sign(&output, stamp_path);

  REQUIRE(written > 0);

  auto ret = std::string(output, written);
  free(output);
  delete doc;

  return ret;
}

Document* CreateNewSignedDocObj(std::string doc_path) {
  auto signedDocBinary = CreateDocSigned(doc_path);

  return fromBuffer(signedDocBinary);
}

Document* fromBuffer(std::string buffer) {
  auto doc = new Document();
  doc->LoadFromBuffer(buffer.c_str(), buffer.length(), "");
  return doc;
}

Document* fromBufferSigned(std::string buffer) {
  auto randnum = rand() % 9000 + 1000;
  std::ofstream write;

  auto filename = "/tmp/" + std::to_string(randnum) +".pdf";
  std::ofstream out(filename);
  out << buffer;
  out.close();

  auto doc = new Document();

  doc->Load(filename, "");

  auto signer = Signer(*doc, "");
  auto field = createField(doc);
  signer.SignatureField = field;

  REQUIRE_NOTHROW(signer.LoadPairFromDisk(
    keys_path + "certificate.pem", keys_path + "private-key.pem", ""));

  char* output = {};
  output = (char*)malloc(650000);
  auto written = signer.sign(&output, stamp_path);


  auto newDoc = new Document();
  
  newDoc->LoadFromBuffer(output,written, "");
  
  free(output);
  return newDoc;
}

Document* ReSignDoc(std::string buffer) {
  auto doc = new Document();

  doc->LoadFromBuffer(buffer.c_str(), buffer.length(), "");
  auto signer = Signer(*doc, "");
  auto field = createField(doc);
  signer.SignatureField = field;

  REQUIRE_NOTHROW(signer.LoadPairFromDisk(
    keys_path + "certificate.pem", keys_path + "private-key.pem", ""));

  char* output = {};
  output = (char*)malloc(650000);
  auto written = signer.sign(&output, stamp_path);

  REQUIRE(written > 0);

  auto ret = std::string(output, written);
  free(output);

  return fromBuffer(ret);
}