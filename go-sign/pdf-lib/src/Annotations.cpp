#include "Annotations.hpp"
#include <iostream>
#include <openssl/err.h>

std::string getOpenSSLError() {
  BIO* bio = BIO_new(BIO_s_mem());
  ERR_print_errors(bio);
  char* buf;
  size_t len = BIO_get_mem_data(bio, &buf);
  std::string ret(buf, len);
  BIO_free(bio);
  return ret;
}

void print_hex(const char* string, std::string start) {
  unsigned char* p = (unsigned char*)string;

  for (int i = 0; i < strlen(string); ++i) {
    if (!(i % 16) && i)
      std::cout << std::endl << start;
    auto v = p[i];
    char print[6];
    sprintf(print, "0x%02x ", v);
    std::cout << print;
  }
  std::cout << std::endl;
}

ASN1_UTCTIME* get_signed_time(PKCS7_SIGNER_INFO* si) {
  ASN1_TYPE* so;
  so = PKCS7_get_signed_attribute(si, NID_pkcs9_signingTime);
  if (so->type == V_ASN1_UTCTIME)
    return so->value.utctime;
  return NULL;
}

Annotations::Annotations(BaseDocument* doc)
  : originalDoc(doc)
  , objs(doc->GetDocument().GetObjects()) {

  for (auto i : objs) {
    try {
      auto dict = i->GetDictionary();
      for (auto key : dict.GetKeys()) {
        if (key.first.GetEscapedName() == "Annots") {
          auto contents = key.second->GetArray();
          for (auto value : contents) {
            auto ref = value.GetReference();
            auto obj = objs.GetObject(ref);
            if (obj->GetDataType() == ePdfDataType_Dictionary) {
              // decodeDictPrint(obj->GetDictionary(), "");
              decodeDict(obj->GetDictionary());
            }
          }
        }
      }
    } catch (...) {
    }
  }

  for (auto& v : vt) {
    verifyAndExtractSig(v);
  }
}

std::string Annotations::sha256(char* buff, int bufflen) {
  unsigned char hash[SHA256_DIGEST_LENGTH];
  SHA256_CTX sha256;
  SHA256_Init(&sha256);
  SHA256_Update(&sha256, buff, bufflen);
  SHA256_Final(hash, &sha256);
  std::stringstream ss;
  for (int i = 0; i < SHA256_DIGEST_LENGTH; i++) {
    ss << std::hex << std::setw(2) << std::setfill('0') << (int)hash[i];
  }
  return ss.str();
}

std::vector<int> Annotations::GetPendingByteRange() {
  for (auto v : vt) {
    if (v.pending) {
      return v.ByteRange;
    }
  }
  return std::vector<int>();
}

std::string Annotations::GetEdgeSignatureHash(edge edge) {
  std::cout << "size " << vt.size() << std::endl;
  if (vt.empty()) {
    return "";
  }

  if (vt.size() == 1) {

    auto len = vt[0].Signature.messageLength;
    auto hash = vt[0].Signature.messageHash;
    std::cout << "msg " << std::string(reinterpret_cast<char*>(hash), len)
              << std::endl;

    return std::string(reinterpret_cast<char*>(hash), len);
  }

  VType* last = nullptr;
  for (auto i = 0; i < vt.size() - 1; i++) {
    auto curr = *vt[i].Signature.timestamp;
    auto next = *vt[i + 1].Signature.timestamp;

    switch (ASN1_TIME_compare(curr, next)) {
      case -1: {
        // case curr is older than next
        auto pos = edge == edge::first ? i : i + 1;
        last = &vt[pos];
        break;
      }

      case 1: {
        // case next is older than curr
        auto pos = edge == edge::first ? i + 1 : i;
        last = &vt[pos];
        break;
      }

      case -2:
        continue;
    }
  }

  if (last == nullptr) {
    return "";
  }

  auto len = last->Signature.messageLength;
  auto hash = reinterpret_cast<char*>(last->Signature.messageHash);

  return std::string(hash, len);
}

void Annotations::verifyAndExtractSig(VType& value) {
  BIO* cont = NULL;

  BIO* in = BIO_new(BIO_s_mem());

  if (!BIO_write(in, value.Contents, value.contentsLen)) {
    throw std::runtime_error("Could not read signature");
  };

  if (value.contentsLen > 5) {
    auto contentStr = std::string(value.Contents, value.Contents + 12);
    const std::string placeHolder = { '\xee', '\xbe', '\xae', '\xee',
                                      '\xbe', '\xae', '\xee', '\xbe',
                                      '\xae', '\xee', '\xbe', '\xae' };
    if (contentStr == placeHolder) {
      value.pending = true;
    }
  }

  value.Signature.p7 = PKCS7_new();
  auto newP7 = d2i_PKCS7_bio(in, &value.Signature.p7);
  if (!newP7) {
    return;
  }

  auto p7 = value.Signature.p7;

  // We'll only allow pkcs7_signed type messages
  // cuz they're the ones we print into the pdf file
  int i = OBJ_obj2nid(p7->type);
  if (i == NID_pkcs7_signed) {
    value.Signature.certs = p7->d.sign->cert;
  } else {
    throw std::runtime_error("Invalid ASN1 type");
  }

  // uncomment following lines to dump P7 data
  // BIO* bio_out = BIO_new_file("dumpannot.txt", "w");

  // BIO* bio_out = BIO_new_fp(stdout, BIO_NOCLOSE);
  // PKCS7_print_ctx(bio_out, p7, 0, NULL);
  // BIO_free(bio_out);

  // ========= end dump =========

  // get message hash from signer info
  auto signerInfos = PKCS7_get_signer_info(p7);
  auto signer_num = sk_PKCS7_SIGNER_INFO_num(signerInfos);
  for (auto i = 0; i < 1; i++) {
    auto si = sk_PKCS7_SIGNER_INFO_value(signerInfos, i);
    auto attribs = PKCS7_get_signed_attributes(si);
    if (attribs == NULL) {
      throw std::runtime_error("missing signed attributes");
    }

    auto asn1Str = PKCS7_digest_from_attributes(attribs);
    value.Signature.messageHash =
      (unsigned char*)calloc(asn1Str->length, sizeof(unsigned char));

    value.Signature.messageLength = asn1Str->length;
    auto t = std::make_shared<ASN1_UTCTIME*>(get_signed_time(si));

    value.Signature.timestamp = t;
    memcpy(value.Signature.messageHash, asn1Str->data, asn1Str->length);
  }

  if (in != nullptr) {
    BIO_free(in);
  }
}

int Annotations::getSignature(char** output) {

  Verify();

  if (vt.size() == 0) {
    return 0;
  }

  auto hash = this->vt.back().Signature.messageHash;
  auto len = this->vt.back().Signature.messageLength;

  if (hash == nullptr) {
    return 0;
  }

  memcpy(*output, hash, len);

  return len;
}

bool Annotations::Verify() {
  if (vt.empty()) {
    return false;
  }

  for (auto v : vt) {
    BIO* signedData = BIO_new(BIO_s_mem());

    originalDoc->GetDocumentSliced(v.ByteRange, signedData);
    auto r = PKCS7_verify(v.Signature.p7,
                          v.Signature.certs,
                          NULL,
                          signedData,
                          nullptr,
                          PKCS7_DETACHED | PKCS7_NOVERIFY);
    if (r == 0) {
      std::cout << getOpenSSLError() << std::endl;
      return false;
    }
    BIO_free(signedData);
  }

  return true;
}

void Annotations::decodeDict(PdfDictionary& dict) {
  auto T = dict.GetKey("T");

  // if not a signature, then just skip it
  if (T == nullptr) {
    return;
  }

  std::cout << T->GetString().GetString() << std::endl;

  if (!strcmp(T->GetString().GetString(), "pmesp.signer")) {
    // allow only  "pmsigner.sign" (our application) to be verified
    auto Vref = dict.GetKey("V");
    auto objValue = objs.GetObject(Vref->GetReference());
    if (objValue->GetDataType() != ePdfDataType_Dictionary) {
      throw std::runtime_error(
        "It was not possible to get value dict from annotation");
    }

    VType vtValue;

    auto VMap = objValue->GetDictionary();

    // extract byterange from a signature annotation object
    std::vector<int> byteRange;
    auto brArray = VMap.GetKey("ByteRange");
    if (brArray->GetDataType() != ePdfDataType_Array) {
      throw std::runtime_error("It was not possible to get byte range info");
    }

    for (auto elem : brArray->GetArray()) {
      if (elem.GetDataType() != ePdfDataType_Number) {
        throw std::runtime_error("Invalid byte range value");
      }
      vtValue.ByteRange.push_back(elem.GetNumber());
    }

    auto contents = VMap.GetKey("Contents");
    if (contents->GetDataType() != ePdfDataType_HexString) {
      throw std::runtime_error("It was not possible to get signature digest");
    }

    // contents is the p7 signature, this is the buffer we want to parse using
    // openssl
    auto contentStr = contents->GetString();
    auto len = contentStr.GetBuffer().GetSize();
    vtValue.Contents = (char*)malloc(len * sizeof(char));

    auto cont = contentStr.GetBuffer().GetBuffer();
    memcpy(vtValue.Contents, cont, len);
    vtValue.contentsLen = len;

    // extract some other useful fata from the dictionary
    // letting this here not because we could need it
    // but to show how the process is made in podofo
    // this is a rare information to find and i've spent some days on this.
    // enjoy

    auto location = VMap.GetKey("Location");
    if (location->GetDataType() != ePdfDataType_String) {
      throw std::runtime_error("It was not possible to get location string");
    }
    vtValue.Location = location->GetString().GetStringUtf8();

    auto Prop_Build = VMap.GetKey("Prop_Build");
    if (Prop_Build->GetDataType() != ePdfDataType_Dictionary) {
      throw std::runtime_error("Missing prop_Build");
    }

    auto Reason = VMap.GetKey("Reason");
    if (Reason->GetDataType() != ePdfDataType_String) {
      throw std::runtime_error("Invalid reason field");
    }

    vtValue.Reason = Reason->GetString().GetStringUtf8();

    auto SubFilter = VMap.GetKey("SubFilter");
    if (SubFilter->GetDataType() != ePdfDataType_Name) {
      throw std::runtime_error("Invalid SubFilter field");
    }
    vtValue.SubFilter = SubFilter->GetName().GetEscapedName();

    auto Type = VMap.GetKey("Type");
    if (Type->GetDataType() != ePdfDataType_Name) {
      throw std::runtime_error("Invalid type field");
    }

    auto AltName = VMap.GetKey("Reason");
    if (!AltName || AltName->GetDataType() != ePdfDataType_String) {
      throw std::runtime_error("Could not find document id");
    }

    vtValue.DocumentID = AltName->GetString().GetString();
    vtValue.Type = Type->GetName().GetEscapedName();
    vt.push_back(vtValue);
  }
}

void Annotations::decodeDictPrint(PdfDictionary& dict, std::string start) {
  start += "\t";
  for (auto pair : dict.GetKeys()) {

    std::cout << start << pair.first.GetEscapedName() << " "
              << pair.second->GetDataTypeString() << std::endl;
    switch (pair.second->GetDataType()) {
      case ePdfDataType_String:

        std::cout << start + "\t" << pair.second->GetString().GetString()
                  << std::endl;
        break;
      case ePdfDataType_Name:
        std::cout << start + "\t" << pair.second->GetName().GetEscapedName()
                  << std::endl;
        break;
      case ePdfDataType_Dictionary:
        decodeDictPrint(pair.second->GetDictionary(), start);
        std::cout << std::endl;
        break;
      case ePdfDataType_Number:
        std::cout << start + "\t" << pair.second->GetNumber() << std::endl;
        break;
      case ePdfDataType_Array:
        decodeArrayPrint(pair.second->GetArray(), start);
        std::cout << std::endl;
        break;
      case ePdfDataType_Reference:
        decodeRefPrint(pair.second->GetReference(), start);
        std::cout << std::endl;
        break;
      case ePdfDataType_HexString:
        std::cout << start + "\t";
        auto hex = pair.second->GetString().GetStringUtf8().c_str();
        print_hex(hex, start + "\t");
        std::cout << std::endl;
        break;
    }
  }
}

void Annotations::decodeRefPrint(PdfReference ref, std::string start) {
  auto obj = objs.GetObject(ref);
  switch (obj->GetDataType()) {
    case ePdfDataType_String:
      std::cout << start + "\t" << obj->GetString().GetString() << std::endl;
      break;
    case ePdfDataType_Name:
      std::cout << start + "\t" << obj->GetName().GetEscapedName() << std::endl;
      break;
    case ePdfDataType_Dictionary:
      decodeDictPrint(obj->GetDictionary(), start);
      std::cout << std::endl;
      break;
    case ePdfDataType_Number:
      std::cout << start + "\t" << obj->GetNumber() << std::endl;
      break;
    case ePdfDataType_Array:
      decodeArrayPrint(obj->GetArray(), start);
      break;
  }
}

void Annotations::decodeArrayPrint(PdfArray& arr, std::string start) {
  for (auto i : arr) {
    switch (i.GetDataType()) {
      case ePdfDataType_String:
        std::cout << start + "\t" << i.GetString().GetString() << std::endl;
        break;
      case ePdfDataType_Name:
        std::cout << start + "\t" << i.GetName().GetEscapedName() << std::endl;
        break;
      case ePdfDataType_Dictionary:
        decodeDictPrint(i.GetDictionary(), start);
        std::cout << "\n";
        break;
      case ePdfDataType_Number:
        std::cout << start + "\t" << i.GetNumber() << std::endl;
        break;
    }
  }
}

bool Annotations::validateID(std::string id) {
  if (vt.empty()) {
    return true;
  } else {
    return id == vt[0].DocumentID;
  }
}

std::string Annotations::getID() {
  return vt.empty() ? "" : vt[0].DocumentID;
}

Annotations::~Annotations() {}
