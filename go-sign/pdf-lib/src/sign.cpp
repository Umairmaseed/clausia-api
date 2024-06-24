#include <Document.hpp>
#include <Signer.hpp>

#include <podofo/podofo-base.h>
#include <podofo/podofo.h>

#include <fstream>
#include <iostream>
#include <streambuf>
#include <string>
#include <type_traits>

#include <openssl/ecdsa.h>
#include <openssl/evp.h>

#include <openssl/pem.h>

using namespace PoDoFo;

void usage() {
  std::cerr << "Usage: \n" << std::endl;
  std::cerr << "\n\tpdf-sign public_key.pem private_key.pem inputfile.pdf "
               "outputfile.pdf\n"
            << std::endl;
}

int main(int argc, char const* argv[]) {
  if (argc == 2 && std::string(argv[1]) == "-h") {
    usage();
    return 0;
  }

  if (argc != 5) {
    std::cerr << "Wrong number of arguments: " << std::endl;
    usage();
    return 1;
  }
  const std::string public_key_path = argv[1];
  const std::string private_key_path = argv[2];
  const std::string input_file = argv[3];
  const std::string output_file = argv[4];
  auto doc = new Document();

  try {
    doc->Load(input_file, "");
    auto rect = PoDoFo::PdfRect(0, 0, 10, 10);
    auto page =
      doc->GetDocument().GetPage(doc->GetDocument().GetPageCount() - 1);
    auto annot = page->CreateAnnotation(
      PoDoFo::EPdfAnnotation::ePdfAnnotation_Widget, rect);
    annot->SetFlags(PoDoFo::EPdfAnnotationFlags::ePdfAnnotationFlags_Hidden |
                    PoDoFo::EPdfAnnotationFlags::ePdfAnnotationFlags_Invisible |
                    PoDoFo::EPdfAnnotationFlags::ePdfAnnotationFlags_Print);
    auto field =
      std::shared_ptr<PoDoFo::PdfSignatureField>(new PoDoFo::PdfSignatureField(
        annot, doc->GetDocument().GetAcroForm(), &doc->GetDocument()));
    field->SetSignatureReason(PoDoFo::PdfString("razao"));
    field->SetSignatureLocation(PoDoFo::PdfString("here"));
    field->SetSignatureCreator(PoDoFo::PdfName("pmesp.signer"));
    field->SetFieldName("pmsigner.sign");
    field->SetSignatureDate(PoDoFo::PdfDate());

    auto signer = Signer(*doc, output_file);
    signer.SignatureField = field;

    signer.LoadPairFromDisk(public_key_path, private_key_path, "");
    auto stamp_path = "./stamps/goledger-icon";

    signer.sign(nullptr, stamp_path);
  } catch (PdfError& e) {
    std::cout << e.ErrorMessage(e.GetError()) << std::endl;
  }

  return 0;
}