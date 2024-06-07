#include "wrapper.hpp"

int writeSignature(char* pdfFile,
                   int fileLength,
                   SignatureParams params,
                   char* stampPath,
                   char* publicKey,
                   char* privateKey,
                   char* password,
                   QrCode qrCode,
                   DetailsInfo di,
                   char* output);

int verifySignature(char* pdfFile, int fileLength, char* password);

int getLastSignature(char* pdfFile, int fileLenght, char* hashout);

int getFirstSignature(char* pdfFile, int fileLen, char* hashout);

int getDocID(char* pdfFile, int fileLength, char* idOut);
