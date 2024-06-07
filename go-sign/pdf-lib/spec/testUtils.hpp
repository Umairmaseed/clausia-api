#ifndef __TESTUTILS__H__
#define __TESTUTILS__H__

#include "Document.hpp"
#include "Signer.hpp"
#include <catch2/catch.hpp>
#include <string>
#include <test_const.h>

std::string CreateDocSigned(std::string doc_path);
Document* CreateNewSignedDocObj(std::string doc_path);
Document* ReSignDoc(std::string buffer);
Document* fromBuffer(std::string buffer);
Document* fromBufferSigned(std::string buffer);

#endif //!__TESTUTILS__H__