#include <atomic>
#include <mutex>
#include <thread>
#include <chrono>

class Limiter {
 public:
  virtual std::chrono::system_clock::time_point take() = 0;
};

// State is a abstract of a limiter state.
struct State {
  State(){};
  State(std::chrono::system_clock::time_point last, std::chrono::duration<int, std::milli> sleep_for) : last_(last), sleep_for_(sleep_for){};

  // last_ represents the last time a request occurs.
  std::chrono::system_clock::time_point last_;

  // sleep_for_ represents how long the thread need to sleep before deal with the next request.
  std::chrono::duration<int, std::milli> sleep_for_;
};

// LimiterAtomic is a rate limiter used atomic operation to avoid race condition.
class LimiterAtomic : public Limiter {
 public:
  LimiterAtomic(int rate, std::chrono::duration<int, std::milli> per, int slack);
  LimiterAtomic(LimiterAtomic &atomic_limiter) = delete;
  LimiterAtomic *operator=(const LimiterAtomic &) = delete;
  std::chrono::system_clock::time_point take() override;

 private:
  // Limiter's state.
  std::atomic<State> state_;

  // Limiter's rate(how long a request should take).
  std::chrono::duration<int, std::milli> per_request_;

  // Limiter's slack(deal with suddent peak flow like token bucket).
  std::chrono::duration<int, std::milli> max_slack_;
};

LimiterAtomic::LimiterAtomic(int rate, std::chrono::duration<int, std::milli> per, int slack)
    : state_(State{std::chrono::system_clock::time_point(), std::chrono::duration<int, std::milli>{0}}),
      per_request_(std::chrono::duration<int, std::milli>{per / std::chrono::duration<int, std::milli>{rate}}),
      max_slack_(-1 * slack * per_request_) {}

std::chrono::system_clock::time_point LimiterAtomic::take() {
  State new_state;
  bool taken = false;
  std::chrono::duration<int, std::milli> interval{0};

  while (!taken) {
    auto now = std::chrono::system_clock::now();
    auto previous_state = state_.load(std::memory_order_relaxed);
    new_state.sleep_for_ = previous_state.sleep_for_;
    new_state.last_ = now;

    if (previous_state.last_ == std::chrono::system_clock::time_point()) {
      taken = state_.compare_exchange_weak(previous_state, new_state, std::memory_order_release, std::memory_order_relaxed);
      continue;
    }

    // sleep_for_ calculates how much time we should sleep based on the per_request_ budget.
    // Since the request may take longer than the budget, sleep_for_ can be negative.
    new_state.sleep_for_ += (per_request_ - std::chrono::duration_cast<std::chrono::milliseconds>(new_state.last_ - previous_state.last_));

    // We shouldn't allow sleep_for_ to get too negative, since it would mean that
    // a service that slowed down a lot for a short period of time would get a much
    // higher RPS following that.
    if (new_state.sleep_for_ < max_slack_)
      new_state.sleep_for_ = max_slack_;

    if (new_state.sleep_for_.count() > 0) {
      new_state.last_ += new_state.sleep_for_;
      interval = new_state.sleep_for_;
      new_state.sleep_for_ = std::chrono::duration<int, std::milli>{0};
    }
    taken = state_.compare_exchange_weak(previous_state, new_state, std::memory_order_release, std::memory_order_relaxed);
  }
  std::this_thread::sleep_for(interval);
  return new_state.last_;
}

// LimiterAtomic is a rate limiter used atomic operation to avoid race condition.
class LimiterMutex : public Limiter {
 public:
  LimiterMutex(int rate, std::chrono::duration<int, std::milli> per, int slack);
  LimiterMutex(LimiterAtomic &atomic_limiter) = delete;
  LimiterMutex *operator=(const LimiterMutex &) = delete;
  std::chrono::system_clock::time_point take() override;

 private:
  std::chrono::system_clock::time_point last_;
  std::chrono::duration<int, std::milli> per_request_;
  std::chrono::duration<int, std::milli> max_slack_;
  std::chrono::duration<int, std::milli> sleep_for_;
  std::mutex mut_;
};

LimiterMutex::LimiterMutex(int rate, std::chrono::duration<int, std::milli> per, int slack)
    : last_(std::chrono::system_clock::time_point()),
      per_request_(std::chrono::duration<int, std::milli>{per / std::chrono::duration<int, std::milli>{rate}}),
      max_slack_(-1 * slack * per_request_),
      sleep_for_(std::chrono::duration<int, std::milli>{0}) {}

std::chrono::system_clock::time_point LimiterMutex::take() {
  std::unique_lock<std::mutex> lock{mut_};
  auto now = std::chrono::system_clock::now();
  if (last_ == std::chrono::system_clock::time_point()) {
    last_ = now;
    return last_;
  }
  sleep_for_ += (per_request_ - std::chrono::duration_cast<std::chrono::milliseconds>(now - last_));

  if (sleep_for_ < max_slack_) sleep_for_ = max_slack_;

  if (sleep_for_.count() > 0) {
    std::this_thread::sleep_for(sleep_for_);
    last_ = now + sleep_for_;
    sleep_for_ = std::chrono::duration<int, std::milli>{0};
  } else {
    last_ = now;
  }
  return last_;
}