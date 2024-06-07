#include "wrapper.hpp"
#include "Document.hpp"
#include "Signer.hpp"
#include <iostream>

int writeSig(char* pdfFile,
             int fileLength,
             SignatureParams params,
             char* stamp_path,
             char* publicKey,
             char* privateKey,
             char* password,
             QrCode qrCode,
             DetailsInfo di,
             char** output) {
  auto doc = new Document();

  try {
    doc->LoadFromBuffer(pdfFile, fileLength, "");
    auto info = doc->GetDocument().GetPage(0)->GetPageSize();

    auto pagenum = params.page < doc->GetDocument().GetPageCount()
                     ? params.page - 1
                     : doc->GetDocument().GetPageCount() - 1;
    if (!params.final) {
      if (params.x < 0) {
        params.x = 0;
      }

      if (params.y < 0) {
        params.y = 0;
      }

    } else {
      params.x = 10;
      params.y = 10;
      pagenum = doc->GetDocument().GetPageCount() - 1;
    }


    auto page = doc->GetDocument().GetPage(pagenum);
    auto pageWidth = page->GetPageSize().GetWidth();
    if (params.x > pageWidth - 180) {
      params.x = pageWidth - 180;
    }

    auto sigRect = PoDoFo::PdfRect(params.x, params.y, 300, 100);
    auto lastSig = doc->GetLastSignature();

    auto id = std::string(qrCode.id, strlen(qrCode.id));

    if (lastSig.empty() && qrCode.qrCode != nullptr) {
      doc->DrawQrCode(QrCodeInfo{
        url : std::string(qrCode.url, strlen(qrCode.url)),
        id : std::string(qrCode.id, strlen(qrCode.id)),
        pData : reinterpret_cast<unsigned char*>(qrCode.qrCode),
        dwLen : qrCode.len,
      },
                      DrawInfo{
                        Re : std::string(di.Re, strlen(di.Re)),
                        Name : std::string(di.Name, strlen(di.Name)),
                        Rank : std::string(di.Rank, strlen(di.Rank))
                      });
    }

    auto annot = page->CreateAnnotation(
      PoDoFo::EPdfAnnotation::ePdfAnnotation_Widget, sigRect);

    annot->SetFlags(PoDoFo::EPdfAnnotationFlags::ePdfAnnotationFlags_Print);
    
    auto field =
      std::shared_ptr<PoDoFo::PdfSignatureField>(new PoDoFo::PdfSignatureField(
        annot, doc->GetDocument().GetAcroForm(), &doc->GetDocument()));

    field->SetSignatureReason(PoDoFo::PdfString(id));
    field->SetSignatureLocation(PoDoFo::PdfString(params.location));
    field->SetSignatureCreator(PoDoFo::PdfName(params.creator));
    field->SetFieldName("pmesp.signer");
    field->SetSignatureDate(PoDoFo::PdfDate());

    auto signer = Signer(*doc, "");
    signer.SignatureField = field;

    signer.LoadPairFromMemory(publicKey, privateKey, password);

    auto writen = signer.sign(output, stamp_path);

    delete doc;

    return writen;
  } catch (const std::exception& e) {
    std::cerr << e.what() << '\n';
    return -1;
  }
  return 0;
}

int verify(char* pdfFile, int fileLength, char* password) {

  auto doc = new Document();
  try {
    doc->LoadFromBuffer(pdfFile, fileLength, password);
    auto valid = doc->Verify();
    if (valid) {
      return 1;
    }
    return 0;
  } catch (const std::exception& e) {
    std::cerr << e.what() << '\n';
    return -1;
  }
  return 0;
}

int getLastSig(char* pdfFile, int fileLen, char** hashout) {
  auto doc = new Document();
  try {
    doc->LoadFromBuffer(pdfFile, fileLen, "");

    auto sig = doc->GetLastSignature();
    memcpy(*hashout, sig.c_str(), sig.length());

    return sig.length();
  } catch (const std::exception& e) {
    std::cerr << e.what() << '\n';
    return -1;
  }
}

int getFirstSig(char* pdfFile, int fileLen, char** hashout) {
  auto doc = new Document();
  try {
    doc->LoadFromBuffer(pdfFile, fileLen, "");

    auto sig = doc->GetFirstSignature();
    memcpy(*hashout, sig.c_str(), sig.length());

    return sig.length();
  } catch (const std::exception& e) {
    std::cerr << "[lib]" << e.what() << '\n';
    return -1;
  }
}

int getID(char* pdfFile, int fileLength, char** idOut) {
  auto doc = new Document();
  try {
    doc->LoadFromBuffer(pdfFile, fileLength, "");

    auto sig = doc->getID();
    memcpy(*idOut, sig.c_str(), sig.length());

    return sig.length();
  } catch (const std::exception& e) {
    std::cerr << e.what() << '\n';
    return -1;
  }
}