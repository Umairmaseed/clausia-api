#include "Document.hpp"

#include <iostream>

Document::Document()
  : BaseDocument(originalBuffer) {}

Document::~Document() {}

void Document::Load(string path, string password) {
  std::ifstream file(path, std::ios::binary | std::ios::ate);
  std::streamsize size = file.tellg();
  file.seekg(0, std::ios::beg);

  std::vector<char> buffer(size);
  if (file.read(buffer.data(), size)) {
    LoadFromBuffer(&buffer[0], size, password);
  } else {
    throw std::invalid_argument("can't read requested file " + path);
  }
}

void Document::LoadFromBuffer(const char* buffer,
                              size_t buffer_len,
                              string password) {
  originalRef = std::vector<char>(buffer, buffer + buffer_len);

  auto device = new PdfRefCountedInputDevice(buffer, buffer_len);

  load("", device, password, DocumentInputDevice::Memory);
  annots = new Annotations(static_cast<BaseDocument*>(this));
}

std::vector<int> Document::PendingByteRange() {
  return annots->GetPendingByteRange();
}

bool Document::Verify() {
  return annots->Verify();
}

int Document::getSignature(char** output) {
  return annots->getSignature(output);
}

std::string Document::GetLastSignature() {
  return annots->GetEdgeSignatureHash(edge::last);
}

std::string Document::GetFirstSignature() {
  return annots->GetEdgeSignatureHash(edge::first);
}

void Document::DrawQrCode(QrCodeInfo qci, DrawInfo di) {

  PdfPainter painter;
  auto nPages = GetDocument().GetPageCount();
  auto page = GetDocument().GetPage(nPages - 1);
  auto size = page->GetPageSize();
   unsigned int width = 500;
  unsigned int height = size.GetHeight() - 40;

  PdfRect pdfRect(450, 0, size.GetWidth()-450, height); 

  auto annot = page->CreateAnnotation(
    PoDoFo::EPdfAnnotation::ePdfAnnotation_Stamp, pdfRect);

  annot->SetFlags(PoDoFo::EPdfAnnotationFlags::ePdfAnnotationFlags_Print);

  auto doc = &GetDocument();

  PdfXObject xObj(pdfRect, doc);
  painter.SetPage(&xObj);
  painter.SetClipRect(pdfRect);
  painter.Stroke();
  painter.Save();
  painter.SetColor(221.0 / 255.0, 228.0 / 255.0, 1.0);

  painter.Restore();
  const PdfEncoding* pEncoding = new PdfIdentityEncoding();
  PdfFont* font = doc->CreateFont("Roboto Thin");
  font->SetFontCharSpace(0.0);
  font->SetFontSize(7);
  painter.SetTransformationMatrix(0, 1, -1, 0, size.GetWidth()+440, 40);

  // Do the drawing
  painter.SetFont(font);
  try {
    PdfImage image(doc);
    image.LoadFromPngData(qci.pData, qci.dwLen);
    auto xMidRect = pdfRect.GetBottom();
    auto yMidRect = pdfRect.GetLeft();

    auto scale = 0.2;
    painter.DrawImage(xMidRect, yMidRect, &image, scale, scale);

    auto xSideImage = xMidRect + (image.GetWidth()* scale) + 10 ;
    auto yTextStart = yMidRect + (image.GetHeight() * scale) - 30;
   
    auto nowStr = NowStr();


    painter.BeginText(xSideImage, yTextStart);
    painter.AddText(PdfString(reinterpret_cast<const pdf_utf8*>(
      ("Assinatura eletrônica avançada nos termos da Lei nº 14.063/20 por " +
       di.Rank + " " + di.Re + " " + di.Name + " " + nowStr)
        .c_str())));
    painter.EndText();

    yTextStart -= 10;

    painter.BeginText(xSideImage, yTextStart);
    painter.AddText(PdfString(reinterpret_cast<const pdf_utf8*>(
      ("Para conferir a autenticidade, acesse o site " + qci.url + qci.id).c_str())));
    painter.EndText();

  } catch (PdfError& e) {
    std::cout << e.what() << std::endl;
  };

  painter.FinishPage();

  annot->SetAppearanceStream(&xObj);
}

bool Document::IsValidID(std::string id) {
  return annots->validateID(id);
}

std::string Document::getID() {
  return annots->getID();
}