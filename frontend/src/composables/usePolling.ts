import { onBeforeUnmount, ref } from 'vue'

export function usePolling(callback: () => void | Promise<void>, interval = 10000) {
  const timer = ref<ReturnType<typeof setInterval> | null>(null)

  const stop = () => {
    if (timer.value) {
      clearInterval(timer.value)
      timer.value = null
    }
  }

  const start = () => {
    stop()
    timer.value = setInterval(() => {
      void callback()
    }, interval)
  }

  onBeforeUnmount(stop)

  return {
    start,
    stop,
  }
}
