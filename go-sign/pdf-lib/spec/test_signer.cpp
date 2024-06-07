#include <catch2/catch.hpp>

#include "Document.hpp"
#include "Signer.hpp"
#include <fstream>
#include <podofo/podofo.h>
#include <string>
#include <testUtils.hpp>
#include <test_const.h>

TEST_CASE("Load keys from disk no password") {

  std::string outputFile = "/tmp/Signer_test_output";
  auto doc = new Document();

  auto signer = new Signer(*doc, outputFile);
  REQUIRE_NOTHROW(signer->LoadPairFromDisk(
    keys_path + "certificate.pem", keys_path + "private-key.pem", ""));
}

TEST_CASE("Load keys from buffer no password") {
  std::fstream file(keys_path + "certificate.pem");
  REQUIRE(file.is_open());

  std::string pubBuffer = std::string(std::istreambuf_iterator<char>(file),
                                      std::istreambuf_iterator<char>());
  REQUIRE_FALSE(pubBuffer.empty());
  file.close();

  file.open(keys_path + "private-key.pem");
  REQUIRE(file.is_open());

  std::string privBuffer = std::string(std::istreambuf_iterator<char>(file),
                                       std::istreambuf_iterator<char>());
  REQUIRE_FALSE(privBuffer.empty());
  file.close();

  std::string outputFile = "/tmp/Signer_test_output";
  auto doc = new Document();

  auto signer = new Signer(*doc, outputFile);
  REQUIRE_NOTHROW(signer->LoadPairFromMemory(pubBuffer, privBuffer, ""));
  REQUIRE(signer->isPairLoaded());
}

TEST_CASE("Sign clean doc") {
  auto docBuf = CreateDocSigned("blank.pdf");
}

TEST_CASE("Sign doc") {
  auto docBuf = CreateDocSigned("blank_signed.pdf");
}

TEST_CASE("Sign doc with write param") {
  CreateDocSigned("new.pdf");
}

TEST_CASE("Get alternate_name from signed doc") { 
  auto doc = CreateNewSignedDocObj("blank.pdf");
}