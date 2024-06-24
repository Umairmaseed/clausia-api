#include "pdfsign.h"
#include <stdio.h>

int writeSignature(char* pdfFile,
                   int fileLength,
                   SignatureParams params,
                   char* stamp_path,
                   char* publicKey,
                   char* privateKey,
                   char* password,
                   QrCode qrCode,
                   DetailsInfo di,
                   char* output) {
  int written = writeSig(pdfFile,
                         fileLength,
                         params,
                         stamp_path,
                         publicKey,
                         privateKey,
                         password,
                         qrCode,
                         di,
                         &output);
  return written;
}

int verifySignature(char* pdfFile, int fileLength, char* password) {
  return verify(pdfFile, fileLength, password);
}

int getLastSignature(char* pdfFile, int fileLen, char* hashout) {
  return getLastSig(pdfFile, fileLen, &hashout);
}

int getFirstSignature(char* pdfFile, int fileLen, char* hashout) {
  return getFirstSig(pdfFile, fileLen, &hashout);
}

int getDocID(char* pdfFile, int fileLength, char* idOut) { 
  return getID(pdfFile, fileLength, &idOut);
}
