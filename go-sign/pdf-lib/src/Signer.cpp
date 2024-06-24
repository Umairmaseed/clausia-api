#include "Signer.hpp"

using namespace PoDoFo;
using std::string;

Signer::Signer(Document& document, string output)
  : Doc(document)
  , Output(output) {}

Signer::~Signer() {}

int Signer::sign(char** writeBuffer, string path) {
  // init openssl
#if OPENSSL_VERSION_NUMBER < 0x10100000L
  SSL_library_init();
  OpenSSL_add_all_algorithms();
  ERR_load_crypto_strings();
  ERR_load_PEM_strings();
  ERR_load_ASN1_strings();
  ERR_load_EVP_strings();
#else
  OPENSSL_init_ssl(0, nullptr);
  OPENSSL_init();
#endif

  PdfObject* acroform = Doc.GetDocument().GetAcroForm(false)->GetObject();
  size_t sigBuffer = 65535, sigBufferLen;
  int rc;
  char* sigData;
  char* outBuffer;
  long outBufferLen;
  BIO* memory;
  BIO* out;
  PKCS7* p7;

  // Set SigFlags in AcroForm as Signed in AppendOnly mode (3)

  if (acroform->GetDictionary().HasKey(PdfName(Name::SIG_FLAGS))) {
    acroform->GetDictionary().RemoveKey(PdfName(Name::SIG_FLAGS));
  }
  pdf_int64 signedAppendModeFlag = 3;
  acroform->GetDictionary().AddKey(PdfName(Name::SIG_FLAGS),
                                   PdfObject(signedAppendModeFlag));

  // Create an output device for the signed document
  PdfRefCountedBuffer Buffer;
  PdfOutputDevice outputDevice = PdfOutputDevice(&Buffer);
  PdfSignOutputDevice signer(&outputDevice);
  // minimally ensure signature field name property is filled, defaults to
  // "NoPoDoFo.SignatureField"
  if (SignatureField->GetFieldName().GetStringUtf8().empty()) {
    SignatureField->SetFieldName("NoPoDoFo.SignatureField");
  }

  // Set Signing Date if empty
  if (!SignatureField->GetSignatureObject()->GetDictionary().HasKey(Name::M)) {
    PdfDate now;
    PdfString str;
    now.ToString(str);

    SignatureField->SetSignatureDate(now);
  }

  // Set output device to write signature to designated area.
  signer.SetSignatureSize(static_cast<size_t>(MinSigSize));

  SignatureField->SetSignature(*signer.GetSignatureBeacon());
  draw(path);

  PdfMemDocument* getDoc = &Doc.GetDocument();
  getDoc->WriteUpdate(&signer, true);
  if (!signer.HasSignaturePosition()) {
    throw std::out_of_range(
      "Cannot find signature position in the document data");
    return 0;
  }

  signer.AdjustByteRange();
  signer.Seek(0);

  // Create signature
  while (static_cast<void>(sigData = reinterpret_cast<char*>(
                             podofo_malloc(sizeof(char) * sigBuffer))),
         !sigData) {
    sigBuffer = sigBuffer / 2;
    if (!sigBuffer) {
      break;
    }
  }

  if (!sigData) {
    throw std::overflow_error("PdfError: Out of Memory.");
    return 0;
  }

  memory = BIO_new(BIO_s_mem());
  if (!memory) {
    podofo_free(sigData);
    throw std::runtime_error("Failed to create input BIO");
    return 0;
  }

  auto inputData = BIO_new(BIO_s_mem());
  if (!inputData) {
    podofo_free(sigData);
    throw std::runtime_error("Failed to create input BIO");
    return 0;
  }

  auto updatedDoc = new Document();

  updatedDoc->LoadFromBuffer(Buffer.GetBuffer(), Buffer.GetSize(), "");
  auto br = updatedDoc->PendingByteRange();

  int written = updatedDoc->GetDocumentSliced(br, inputData);
  if (!br.empty() && !written) {
    throw std::runtime_error("Failed to slice document at signing process");
  }

  p7 =
    PKCS7_sign(Cert, Pkey, nullptr, inputData, PKCS7_DETACHED | PKCS7_BINARY);
  if (!p7) {
    BIO_free(memory);
    podofo_free(sigData);
    throw std::runtime_error("PKCS7 Sign failed");
    return 0;
  }

  while (sigBufferLen = signer.ReadForSignature(sigData, sigBuffer),
         sigBufferLen > 0) {
    rc = BIO_write(memory, sigData, static_cast<int>(sigBufferLen));
    if (static_cast<unsigned int>(rc) != sigBufferLen) {
      PKCS7_free(p7);
      BIO_free(memory);
      podofo_free(sigData);
      throw std::runtime_error("BIO write failed");
    }
  }

  podofo_free(sigData);
  if (PKCS7_final(p7, memory, PKCS7_DETACHED | PKCS7_BINARY) <= 0) {
    PKCS7_free(p7);
    BIO_free(memory);
    throw std::runtime_error("pkcs7 final failed");
  }

  i2d_PKCS7_bio(memory, p7);

  outBufferLen = BIO_get_mem_data(memory, &outBuffer);
  if (outBufferLen > 0 && outBuffer) {
    if (static_cast<size_t>(outBufferLen) > signer.GetSignatureSize()) {
      PKCS7_free(p7);
      BIO_free(memory);
      BIO_free(out);
      throw std::runtime_error("Signature value out of prescribed range");
    }
    PdfData signature(outBuffer, static_cast<size_t>(outBufferLen));
    signer.SetSignature(signature);
    signer.Flush();
    outputDevice.Flush();

    if (writeBuffer != nullptr) {
      memcpy(*writeBuffer, Buffer.GetBuffer(), Buffer.GetSize());
      return Buffer.GetSize();
    }

    return outputDevice.GetLength();
  } else {
    throw std::runtime_error("Invalid Signature was generated");
    return 0;
  }
}

pdf_int32 Signer::minSignerSize(FILE* file) {
  pdf_int32 size = 0;
  if (fseeko(file, 0, SEEK_END) != -1) {
    size += ftello(file);
  } else {
    size += 3072;
  }
  return size;
}

int pkeyCallback(char* buf,
                 int bufSize,
                 int PODOFO_UNUSED_PARAM(rwflag),
                 void* userData) {
  const char* password = static_cast<char*>(userData);
  if (!password) {
    return 0;
  }
  auto res = static_cast<int>(strlen(password));
  if (res > bufSize) {
    res = bufSize;
  }
  memcpy(buf, password, static_cast<size_t>(res));
  return res;
};

void Signer::LoadPairFromDisk(string pub, string priv, std::string password) {
  FILE* file;

  if (!(file = fopen(pub.c_str(), "rb"))) {
    throw std::runtime_error("Could not open public cert at path " + pub);
  }

  Cert = PEM_read_X509(file, nullptr, nullptr, nullptr);
  MinSigSize += minSignerSize(file);

  fclose(file);

  if (!Cert) {
    throw std::runtime_error("Could not load certificate");
  }

  if (!(file = fopen(priv.c_str(), "rb"))) {
    X509_free(Cert);
    throw std::runtime_error("Could not open private key at path " + priv);
  }

  Pkey = PEM_read_PrivateKey(
    file, nullptr, pkeyCallback, static_cast<void*>(&password));
  MinSigSize += minSignerSize(file);

  if (!Pkey) {
    X509_free(Cert);
    throw std::runtime_error("Failed to decode private key file");
  }

  fclose(file);
}


void Signer::draw(string path) {
  auto rect = SignatureField->GetWidgetAnnotation()->GetRect();
  PdfPainter painter;

  unsigned int width = rect.GetWidth();
  unsigned int height = rect.GetHeight();

  auto doc = &Doc.GetDocument();
  PdfRect pdfRect(rect.GetLeft(), rect.GetBottom(), width, height);

  PdfXObject xObj(pdfRect, doc);
  painter.SetPage(&xObj);
  painter.SetClipRect(pdfRect);

  painter.Save();
  painter.SetColor(221.0 / 255.0, 228.0 / 255.0, 1.0);

  painter.Restore();
  const PdfEncoding* pEncoding = new PdfIdentityEncoding();
  PdfFont* font = doc->CreateFont("Roboto Thin");
  font->SetFontCharSpace(0.0);
  font->SetFontSize(8);

  // Do the drawing
  painter.SetFont(font);
  try {
    PdfImage image(doc);
    image.LoadFromPng(path.c_str());
    auto xMidRect = rect.GetLeft();
    auto yMidRect = rect.GetBottom() + (image.GetHeight() / 7);

    painter.DrawImage(xMidRect, yMidRect, &image, 0.12, 0.12);
  } catch (PdfError& e) {
  };

  painter.BeginText(rect.GetLeft() + 50,
                    rect.GetBottom() + rect.GetHeight() - 25 );
  painter.SetStrokeWidth(20);
  painter.AddText(
    PdfString(reinterpret_cast<const pdf_utf8*>("assinado eletronicamente em")));
  painter.EndText();

  auto t = PdfDate().GetTime();
  tm* local = localtime(&t);

  auto hourFormat = [](int hour) {
    auto timezoned = hour - 3;
    if (timezoned < 0) {
      timezoned = 24 - timezoned;
    }

    return numFormat(timezoned);
  };

  painter.BeginText(rect.GetLeft() + 50,
                    rect.GetBottom() + rect.GetHeight() - 35);
  painter.AddText(PdfString(numFormat(local->tm_mday) + "/" +
                            numFormat(1 + local->tm_mon) + "/" +
                            numFormat(1900 + local->tm_year)));
  string hour = " às " + hourFormat(local->tm_hour) + ":" +
                numFormat(local->tm_min) + ":" + numFormat(local->tm_sec) + " por";
  painter.AddText(PdfString(reinterpret_cast<const pdf_utf8*>(hour.c_str())));
  painter.EndText();

  char buf[256];
  X509_NAME_oneline(X509_get_subject_name(Cert), buf, 256);
  auto str = string(buf, 256);
  str.shrink_to_fit();
  str.erase(str.begin(), str.begin() + str.find("CN=") + 3);

  painter.BeginText(rect.GetLeft() + 50,
                    rect.GetBottom() + rect.GetHeight() - 45);
  painter.AddText(PdfString(reinterpret_cast<const pdf_utf8*>(str.c_str())));
  painter.EndText();

  painter.FinishPage();

  SignatureField->SetReadOnly(true);

  PoDoFo::PdfDictionary dict;
  dict.AddKey("N", xObj.GetObject()->Reference());
  SignatureField->GetWidgetAnnotation()->GetObject()->GetDictionary().AddKey(
    "AP", dict);
}

void Signer::LoadPairFromMemory(string pub, string priv, string password) {
  BIO* buffer;
  auto certLength = pub.length();

  buffer = BIO_new_mem_buf(pub.c_str(), certLength);
  Cert = PEM_read_bio_X509(buffer, nullptr, nullptr, nullptr);
  MinSigSize += certLength;

  BIO_free(buffer);
  if (!Cert) {
    throw std::invalid_argument("Could not load certificate");
  }

  // load private key
  buffer = BIO_new(BIO_s_mem());
  auto keyLength = priv.length();
  int length = BIO_write(buffer, priv.c_str(), keyLength);
  Pkey = PEM_read_bio_PrivateKey(
    buffer, nullptr, pkeyCallback, static_cast<void*>(&password));
  MinSigSize += keyLength;

  if (!Pkey) {
    X509_free(Cert);
    throw std::runtime_error("Failed to decode private key file");
  }
  MinSigSize += 100;
}

bool Signer::isPairLoaded() {
  return (Pkey && Cert);
}
