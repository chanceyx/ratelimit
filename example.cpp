#include <iostream>

#include "rate_limiter.hpp"

int main() {
  std::cout << "<--------- atomic limiter --------->" << std::endl;
  LimiterAtomic limiter_atomic{100, std::chrono::seconds(1), 10};
  std::cout << "<---------- test limiter ---------->" << std::endl;
  auto prev = std::chrono::system_clock::now();
  for (int i = 0; i < 10; i++) {
    auto now = limiter_atomic.take();
    if (i > 0) std::cout << i << " " << std::chrono::duration_cast<std::chrono::milliseconds>(now - prev).count() << "ms" << std::endl;
    prev = now;
  }

  std::cout << "<----------- test slack ----------->" << std::endl;
  std::this_thread::sleep_for(std::chrono::seconds(1));
  prev = std::chrono::system_clock::now();
  for (int i = 0; i < 20; i++) {
    auto now = limiter_atomic.take();
    if (i > 0) std::cout << i << " " << std::chrono::duration_cast<std::chrono::milliseconds>(now - prev).count() << "ms" << std::endl;
    prev = now;
  }

  std::cout << "\n";
  std::cout << "<--------- mutex_ limiter --------->" << std::endl;
  LimiterMutex limiter_mutex{100, std::chrono::seconds(1), 10};
  std::cout << "<---------- test limiter ---------->" << std::endl;
  prev = std::chrono::system_clock::now();
  for (int i = 0; i < 10; i++) {
    auto now = limiter_mutex.take();
    if (i > 0) std::cout << i << " " << std::chrono::duration_cast<std::chrono::milliseconds>(now - prev).count() << "ms" << std::endl;
    prev = now;
  }

  std::cout << "<----------- test slack ----------->" << std::endl;
  std::this_thread::sleep_for(std::chrono::seconds(1));
  prev = std::chrono::system_clock::now();
  for (int i = 0; i < 20; i++) {
    auto now = limiter_mutex.take();
    if (i > 0) std::cout << i << " " << std::chrono::duration_cast<std::chrono::milliseconds>(now - prev).count() << "ms" << std::endl;
    prev = now;
  }

  return 0;
}