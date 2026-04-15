export const debounce = function (fn, delay) {
  let timer = null;

  return function (...args) {
    clearTimeout(timer);
    timer = setTimeout(() => fn.apply(this, args), delay);
  };
};

export const createGridAnimator = function (drawFn, initialSettings) {
  let current = { ...initialSettings };
  let queue = [];
  let animating = false;
  let duration = 150; // ms

  const push = (settings) => {
    queue.push(settings);
    process();
  };

  const process = () => {
    if (animating || queue.length === 0) return;
    const next = queue.shift();
    animateTo(next);
  };

  const animateTo = (target) => {
    animating = true;
    const start = performance.now();
    const from = { ...current };

    function frame(now) {
      const t = Math.min((now - start) / duration, 1);

      const blended = {
        offsetX: from.offsetX + (target.offsetX - from.offsetX) * t,
        offsetY: from.offsetY + (target.offsetY - from.offsetY) * t,
        cellW: from.cellW + (target.cellW - from.cellW) * t,
        cellH: from.cellH + (target.cellH - from.cellH) * t,

        // non-numeric fields snap to target
        theme: target.theme,
        snap: target.snap,
        diagonal: target.diagonal,
      };

      drawFn(blended);

      if (t < 1) {
        requestAnimationFrame(frame);
      } else {
        current = target;
        animating = false;
        process();
      }
    }

    requestAnimationFrame(frame);
  };

  return {
    push,
    getCurrent: () => current,
    setDuration: (ms) => (duration = ms),
  };
};
