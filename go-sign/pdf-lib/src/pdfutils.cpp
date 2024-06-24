#include <pdfutils.hpp>

std::string numFormat(int number) {
  return (number < 10 ? "0" : "") + std::to_string(number);
}

std::string NowStr() {
  auto hourFormat = [](int hour) {
    auto timezoned = hour - 3;
    if (timezoned < 0) {
      timezoned = 24 - timezoned;
    }
    return numFormat(timezoned);
  };

  auto t = PoDoFo::PdfDate().GetTime();
  tm* local = localtime(&t);

  auto date = "em " + numFormat(local->tm_mday) + "/" +
                            numFormat(1 + local->tm_mon) + "/" +
                            numFormat(1900 + local->tm_year);
  auto time = " às " + hourFormat(local->tm_hour) + ":" +
              numFormat(local->tm_min) + ":" + numFormat(local->tm_sec);
  
  return date + time;
}