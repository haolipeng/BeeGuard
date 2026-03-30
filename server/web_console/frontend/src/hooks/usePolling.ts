import { useEffect, useRef, useCallback } from 'react';

const DEFAULT_INTERVAL = 30000;

export function usePolling(callback: () => void | Promise<void>, interval: number = DEFAULT_INTERVAL) {
  const savedCallback = useRef(callback);
  const timerRef = useRef<ReturnType<typeof setInterval>>();

  useEffect(() => {
    savedCallback.current = callback;
  }, [callback]);

  const start = useCallback(() => {
    if (timerRef.current) return;
    timerRef.current = setInterval(() => {
      savedCallback.current();
    }, interval);
  }, [interval]);

  const stop = useCallback(() => {
    if (timerRef.current) {
      clearInterval(timerRef.current);
      timerRef.current = undefined;
    }
  }, []);

  useEffect(() => {
    savedCallback.current();
    start();
    return stop;
  }, [start, stop]);

  return { start, stop };
}
