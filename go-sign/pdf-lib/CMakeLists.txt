cmake_minimum_required (VERSION 3.11)

set (CMAKE_CXX_STANDARD_REQUIRED ON)
set (CMAKE_CXX_STANDARD 17)
set (CMAKE_BUILD_TYPE Debug)

# Project name
project (pdf-sign VERSION 1.0 DESCRIPTION "mylib description")


find_path(PODOFO_INCLUDE_DIR podofo/podofo.h
	          PATHS /usr/include /usr/local/include)
find_library(PODOFO_LIBRARY
                NAMES podofo
                PATHS /usr/lib64
                /usr/lib
                /usr/local/lib)


set (SOURCE_DIR src)
set (INCLUDE_DIR 
  include
  ${CMAKE_BINARY_DIR}
  ${CMAKE_SOURCE_DIR}
  ${LIBJPEG_INCLUDE_DIR}
  ${OPENSSL_INCLUDE_DIR}
  ${PODOFO_INCLUDE_DIR}
)


set (SOURCE_FILES src/BaseDocument.cpp  
 src/sign.cpp    
 src/Document.cpp      
 src/Signer.cpp
 src/Annotations.cpp
 src/pdfutils.cpp
)

Include(FetchContent)

FetchContent_Declare(
  Catch2
  GIT_REPOSITORY https://github.com/catchorg/Catch2.git
  GIT_TAG        v2.13.1)

FetchContent_MakeAvailable(Catch2)

find_package(OpenSSL REQUIRED)

message("Creating build output directory: build/${CMAKE_BUILD_TYPE}")
file(MAKE_DIRECTORY ${CMAKE_SOURCE_DIR}/build/${CMAKE_BUILD_TYPE})

# library compilation

add_library(pdfsign SHARED ${SOURCE_FILES} src/wrapper.cpp src/pdfsign.c)

# lib properties
set_target_properties(pdfsign PROPERTIES VERSION ${PROJECT_VERSION})
set_target_properties(pdfsign PROPERTIES SOVERSION 1)
set_target_properties(pdfsign PROPERTIES PUBLIC_HEADER "include/pdfsign.h;include/wrapper.hpp")
set_target_properties(pdfsign PROPERTIES LINKER_LANGUAGE CXX)

# configure
configure_file(pdfsign.pc.in pdfsign.pc @ONLY)

# include to lib
target_include_directories(pdfsign PRIVATE include)
target_include_directories(pdfsign PRIVATE src)

target_link_libraries(pdfsign
                      PRIVATE
                      ${PODOFO_LIBRARY}
                      ${OPENSSL_LIBRARIES}
                      OpenSSL::SSL
                      OpenSSL::Crypto
)
add_compile_options("-lstdc++")
link_libraries("-lstdc++")

# library install
include(GNUInstallDirs)
install(TARGETS pdfsign
    LIBRARY DESTINATION ${CMAKE_INSTALL_LIBDIR}
    PUBLIC_HEADER DESTINATION ${CMAKE_INSTALL_INCLUDEDIR})

install(FILES ${CMAKE_BINARY_DIR}/pdfsign.pc
    DESTINATION ${CMAKE_INSTALL_DATAROOTDIR}/pkgconfig)

# Executable name
add_executable (pdf-sign ${DOMAINS_SOURCE_FILES}
        ${SOURCE_FILES}
)

target_include_directories(pdf-sign 
                           PRIVATE
                           ${INCLUDE_DIR}
                           ${PODOFO_INCLUDE_DIR}
)

target_link_libraries(pdf-sign
                      PRIVATE
                      ${PODOFO_LIBRARY}
                      ${OPENSSL_LIBRARIES}
                      OpenSSL::SSL
                      OpenSSL::Crypto
)


add_custom_command(TARGET pdf-sign POST_BUILD
	                 COMMAND "${CMAKE_COMMAND}" -E copy $<TARGET_FILE:pdf-sign>
	                 ${CMAKE_SOURCE_DIR}/build/${CMAKE_BUILD_TYPE}
)


# build tests
set(testing_sources ${SOURCE_FILES})

# remove main from sources
list(REMOVE_ITEM testing_sources src/sign.cpp)


file (GLOB test_files "spec/*.cpp")
add_executable (pdfSign_test ${testing_sources} ${test_files})

target_include_directories(pdfSign_test 
                           PRIVATE
                           ${INCLUDE_DIR}
                           spec/
                           ${PODOFO_INCLUDE_DIR}
)



list(APPEND CMAKE_MODULE_PATH ${catch2_SOURCE_DIR}/contrib)

target_link_libraries(pdfSign_test
                      ${PODOFO_LIBRARY}
                      ${OPENSSL_LIBRARIES}
                      OpenSSL::SSL
                      Catch2::Catch2
                      OpenSSL::Crypto
)
include(CTest)
include(Catch)
catch_discover_tests(pdfSign_test)

