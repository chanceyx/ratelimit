#include <atomic>
#include <iostream>
#include <memory>
#include <thread>

class Limiter {
 public:
  virtual std::chrono::system_clock::time_point take() = 0;
};

struct State {
  State(){};
  std::chrono::system_clock::time_point last_;
  std::chrono::duration<int, std::milli> sleep_for_;
};

class AtomicLimiter : public Limiter {
 public:
  AtomicLimiter(int rate, std::chrono::duration<int, std::milli> per, int slack);
  std::chrono::system_clock::time_point take() override;

 private:
  std::atomic<State> state_;
  std::chrono::duration<int, std::milli> per_request_;
  std::chrono::duration<int, std::milli> max_slack_;
};

AtomicLimiter::AtomicLimiter(int rate, std::chrono::duration<int, std::milli> per, int slack)
    : state_(State()) {
  auto dur_per = per / std::chrono::duration<int, std::milli>(rate);
  per_request_ = std::chrono::duration<int, std::milli>(dur_per);
  max_slack_ = -1 * slack * per_request_;
  State initial_state;
  initial_state.last_ = std::chrono::system_clock::time_point();
  initial_state.sleep_for_ = std::chrono::duration<int, std::milli>(0);
  state_.store(initial_state, std::memory_order_relaxed);
}

std::chrono::system_clock::time_point AtomicLimiter::take() {
  State new_state;
  bool taken = false;
  std::chrono::duration<int, std::milli> interval(0);

  while (!taken) {
    auto now = std::chrono::system_clock::now();
    auto previous_state = state_.load(std::memory_order_relaxed);
    new_state.sleep_for_ = previous_state.sleep_for_;
    new_state.last_ = now;

    if (previous_state.last_ == std::chrono::system_clock::time_point()) {
      taken = state_.compare_exchange_weak(previous_state, new_state, std::memory_order_release, std::memory_order_relaxed);
      continue;
    }
    new_state.sleep_for_ += (per_request_ - std::chrono::duration_cast<std::chrono::milliseconds>(new_state.last_ - previous_state.last_));

    if (new_state.sleep_for_ < max_slack_)
      new_state.sleep_for_ = max_slack_;

    if (new_state.sleep_for_.count() > 0) {
      new_state.last_ += new_state.sleep_for_;
      interval = new_state.sleep_for_;
      new_state.sleep_for_ = std::chrono::duration<int, std::milli>(0);
    }
    taken = state_.compare_exchange_weak(previous_state, new_state, std::memory_order_release, std::memory_order_relaxed);
  }
  std::this_thread::sleep_for(interval);
  return new_state.last_;
}