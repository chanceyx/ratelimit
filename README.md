# cpp rate limiter

A header-only C++ rate limiter. Based on c++11.

English | [中文](README-zh.md)

## Platforms

- Linux(gcc 5.4.0+)
- maxOS(clang 3.5+)

## Getting started

```shell
$ git clone https://github.com/chanceyx/ratelimit.git
```

copy the `rate_limiter.hpp` file to your build tree, and you're ready to go.

## Examples

```shell
$ git clone https://github.com/chanceyx/ratelimit.git
$ make example
$ ./example
```

## Usage

**atomic based rate limiter usage:**

```c++
#include <iostream>

#include "rate_limiter.hpp"

int main() {
  // rate = 100, duration = 1 second, max_slack = 10
  // Create a atomic based rate limiter with a maximum 100 operations to perform per second.
  LimiterAtomic limiter_atomic(100, std::chrono::seconds(1), 10);
  auto prev = std::chrono::system_clock::now();
  for (int i = 0; i < 10; i++) {
    auto now = limiter_atomic.take();
    if (i > 0) std::cout << i << " " << std::chrono::duration_cast<std::chrono::milliseconds>(now - prev).count() << "ms" << std::endl;
    prev = now;
  }
  return 0;

  // Output:
  // <--------- atomic limiter --------->
  // <---------- test limiter ---------->
  // 1 10ms
  // 2 10ms
  // 3 10ms
  // 4 10ms
  // 5 10ms
  // 6 10ms
  // 7 10ms
  // 8 10ms
  // 9 10ms
}
```

**atomic based slack usage:**

```c++

#include <iostream>

#include "rate_limiter.hpp"

int main() {
  // rate = 100, duration = 1 second, max_slack = 10
  // Create a atomic based rate limiter with a maximum 100 operations to perform per second.
  // A service that slowed down a lot, in a short period it can handle 10 operations
  // with no limit(like token bucket).
  std::cout << "<--------- atomic limiter --------->" << std::endl;
  LimiterAtomic limiter_atomic(100, std::chrono::seconds(1), 10);
  std::cout << "<---------- test limiter ---------->" << std::endl;
  auto prev = std::chrono::system_clock::now();
  for (int i = 0; i < 10; i++) {
    auto now = limiter_atomic.take();
    if (i > 0) std::cout << i << " " << std::chrono::duration_cast<std::chrono::milliseconds>(now - prev).count() << "ms" << std::endl;
    prev = now;
  }

  std::cout << "<----------- test slack ----------->" << std::endl;
  // slow down
  std::this_thread::sleep_for(std::chrono::seconds(1));
  prev = std::chrono::system_clock::now();
  for (int i = 0; i < 20; i++) {
    auto now = limiter_atomic.take();
    if (i > 0) std::cout << i << " " << std::chrono::duration_cast<std::chrono::milliseconds>(now - prev).count() << "ms" << std::endl;
    prev = now;
  }
  return 0;

  // Output
  // <--------- atomic limiter --------->
  // <---------- test limiter ---------->
  // 1 10ms
  // 2 10ms
  // 3 10ms
  // 4 10ms
  // 5 10ms
  // 6 10ms
  // 7 10ms
  // 8 10ms
  // 9 10ms
  // <----------- test slack ----------->
  // 1 0ms
  // 2 0ms
  // 3 0ms
  // 4 0ms
  // 5 0ms
  // 6 0ms
  // 7 0ms
  // 8 0ms
  // 9 0ms
  // 10 0ms
  // 11 10ms
  // 12 10ms
  // 13 10ms
  // 14 10ms
  // 15 10ms
  // 16 10ms
  // 17 10ms
  // 18 10ms
  // 19 10ms
}
```

**mutex based rate limiter usage:**

```c++
#include <iostream>

#include "rate_limiter.hpp"

int main() {
  std::cout << "<--------- mutex_ limiter --------->" << std::endl;
  LimiterMutex limiter_mutex(100, std::chrono::seconds(1), 10);
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

  // Output:
  // <--------- mutex_ limiter --------->
  // <---------- test limiter ---------->
  // 1 10ms
  // 2 10ms
  // 3 10ms
  // 4 10ms
  // 5 10ms
  // 6 10ms
  // 7 10ms
  // 8 10ms
  // 9 10ms
  // <----------- test slack ----------->
  // 1 0ms
  // 2 0ms
  // 3 0ms
  // 4 0ms
  // 5 0ms
  // 6 0ms
  // 7 0ms
  // 8 0ms
  // 9 0ms
  // 10 0ms
  // 11 10ms
  // 12 10ms
  // 13 10ms
  // 14 10ms
  // 15 10ms
  // 16 10ms
  // 17 10ms
  // 18 10ms
  // 19 10ms
}
```

## Reference

uber-go: https://github.com/uber-go/ratelimit
