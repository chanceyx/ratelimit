#include <iostream>

#include "limiter_atomic.hpp"

int main() {
  AtomicLimiter limiter(100, std::chrono::seconds(1), 100);
  auto prev = std::chrono::system_clock::now();
  for (int i = 0; i < 10; i++) {
    auto now = limiter.take();
    if (i > 0) std::cout << i << " " << std::chrono::duration_cast<std::chrono::milliseconds>(now - prev).count() << "ms" << std::endl;
    prev = now;
  }
  return 0;
}