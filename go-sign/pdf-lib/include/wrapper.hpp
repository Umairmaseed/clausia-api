typedef struct SignatureParams {
  char* reason;
  char* location;
  char* creator;
  int x;
  int y;
  int page;
  int final;
} SignatureParams;

typedef struct DetailsInfo {
  char* Name;
  char* Re;
  char* Rank;
} DetailsInfo;

typedef struct QrCode
{
  char* qrCode;
  int len;
  char* id;
  char* url;
} QrCode;

#ifdef __cplusplus
extern "C" {
#endif
int writeSig(char* pdfFile,
             int fileLength,
             SignatureParams params,
             char* stamp_path,
             char* publicKey,
             char* privateKey,
             char* password,
             QrCode qrCode,
             DetailsInfo di,
             char** output);
int verify(char* pdfFile, int fileLength, char* password);
int getLastSig(char* pdfFile, int fileLength, char** signatureHash);
int getFirstSig(char* pdfFile, int fileLen, char** hashout);
int getID(char* pdfFile, int fileLength, char** idOut);

// Your prototype or Definition
#ifdef __cplusplus
}
#endif
