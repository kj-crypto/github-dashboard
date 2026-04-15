import { createWSClient } from './wsClient.js';
import { debounce } from './utils.js';

const ws = createWSClient(`ws://${location.host}`);

const panel = document.getElementById('panel');
const cellWInput = panel.querySelector('#cellW');
const cellHInput = panel.querySelector('#cellH');
const offsetXInput = panel.querySelector('#offsetX');
const offsetYInput = panel.querySelector('#offsetY');
const themeSelect = panel.querySelector('#theme');
const diagonalChbx = panel.querySelector('#diagonal');
const snaplChbx = panel.querySelector('#snap');
const presetSelect = panel.querySelector('#preset');

const status = document.getElementById('status');
const statusDot = status.querySelector('.status-dot');
const statusText = status.querySelector('span:last-child');

ws.onStatusChange((statusMsg) => {
  statusDot.className = `status-dot ${statusMsg}`;
  statusText.textContent = statusMsg;
});

themeSelect.onchange = () => debouncedSendUpdate();
snaplChbx.onchange = () => debouncedSendUpdate();
diagonalChbx.onchange = () => debouncedSendUpdate();

const assignNumberValidation = function (element, min, max) {
  element.min = min;
  element.max = max;

  let lastValidValue = element.defaultValue;

  const isEmpty = (input) => input.value === '' && !input.validity.badInput;
  const isNumber = (input) => input.value !== '' && !input.validity.badInput;
  const isValidNumber = (input) => isNumber(input) && !input.validity.rangeUnderflow && !input.validity.rangeOverflow;

  element.oninput = function () {
    if (isEmpty(this)) {
      this.classList.toggle('invalid', false);
      return;
    }

    if (isValidNumber(this)) {
      this.classList.toggle('invalid', false);
      lastValidValue = this.value;
      debouncedSendUpdate();
    } else if (isNumber(this)) {
      let value;
      if (this.validity.rangeUnderflow) {
        value = min;
      } else if (this.validity.rangeOverflow) {
        value = max;
      }
      this.value = value;
      this.classList.toggle('invalid', true);
      lastValidValue = value;
      debouncedSendUpdate();
    } else {
      this.classList.toggle('invalid', true);
    }
  };

  element.addEventListener('focusout', function () {
    if (!isValidNumber(this)) {
      this.value = lastValidValue;
    }
    this.classList.toggle('invalid', false);
  });
};

assignNumberValidation(cellWInput, 1, 500);
assignNumberValidation(cellHInput, 1, 500);
assignNumberValidation(offsetXInput, -5_000, 5_000);
assignNumberValidation(offsetYInput, -5_000, 5_000);

const collectSettings = function () {
  return {
    offsetX: Number(offsetXInput.value),
    offsetY: Number(offsetYInput.value),
    cellW: Number(cellWInput.value),
    cellH: Number(cellHInput.value),
    theme: themeSelect.value,
    snap: panel.querySelector('#snap').checked,
    diagonal: panel.querySelector('#diagonal').checked,
  };
};

const debouncedSendUpdate = debounce(() => {
  const msg = {
    type: 'update',
    settings: collectSettings(),
  };
  ws.sendJSON(msg);
}, 50);

const presets = {
  kitty: { cellW: 9, cellH: 19 },
  alacritty: { cellW: 9, cellH: 18 },
  iterm: { cellW: 8, cellH: 16 },
  warp: { cellW: 9, cellH: 19 },
};

presetSelect.onchange = () => {
  const preset = presetSelect.value;
  if (preset !== 'custom') {
    const { cellW, cellH } = presets[preset];
    cellWInput.value = cellW;
    cellHInput.value = cellH;
  }
  debouncedSendUpdate();
};
