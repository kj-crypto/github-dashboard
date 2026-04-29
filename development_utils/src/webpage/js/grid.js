import { createGridAnimator } from './utils.js';
import { createWSClient } from './wsClient.js';

const ws = createWSClient(`ws://${location.host}`);
const svg = document.getElementById('grid');
const gridGroup = document.getElementById('grid-lines');
const measureGroup = document.getElementById('measure-group');
const dimGroup = document.getElementById('dimension-group');

const defaultSettings = {
  offsetX: 0,
  offsetY: 0,
  cellW: 10,
  cellH: 20,
  theme: 'dark',
  snap: true,
  diagonal: false,
};

const THEMES = {
  dark: '#444',
  light: '#ccc',
  solarized: '#586e75',
  nord: '#4c566a',
  gruvbox: '#665c54',
};

const createGrid = function () {
  const verticalPool = [];
  const horizontalPool = [];
  const diagonalPool = [];

  return function draw(settings) {
    const width = window.innerWidth;
    const height = window.innerHeight;

    svg.setAttribute('width', width);
    svg.setAttribute('height', height);

    const stroke = THEMES[settings.theme] || '#444';
    const { offsetX, offsetY, cellW, cellH, diagonal } = settings;

    // -----------------------------
    // Vertical lines
    // -----------------------------
    let neededV = Math.ceil((width - offsetX) / cellW);
    for (let i = 0; i < neededV; i++) {
      let line = verticalPool[i];
      if (!line) {
        line = document.createElementNS('http://www.w3.org/2000/svg', 'line');
        verticalPool[i] = line;
        gridGroup.appendChild(line);
      }

      const x = offsetX + i * cellW;

      line.setAttribute('x1', x);
      line.setAttribute('y1', offsetY);
      line.setAttribute('x2', x);
      line.setAttribute('y2', height);
      line.setAttribute('stroke', stroke);
      line.setAttribute('stroke-width', 1);
      line.style.display = 'block';
    }

    // hide unused
    for (let i = neededV; i < verticalPool.length; i++) {
      verticalPool[i].style.display = 'none';
    }

    // -----------------------------
    // Horizontal lines
    // -----------------------------
    let neededH = Math.ceil((height - offsetY) / cellH);
    for (let i = 0; i < neededH; i++) {
      let line = horizontalPool[i];
      if (!line) {
        line = document.createElementNS('http://www.w3.org/2000/svg', 'line');
        horizontalPool[i] = line;
        gridGroup.appendChild(line);
      }

      const y = offsetY + i * cellH;

      line.setAttribute('x1', offsetX);
      line.setAttribute('y1', y);
      line.setAttribute('x2', width);
      line.setAttribute('y2', y);
      line.setAttribute('stroke', stroke);
      line.setAttribute('stroke-width', 1);
      line.style.display = 'block';
    }

    // hide unused
    for (let i = neededH; i < horizontalPool.length; i++) {
      horizontalPool[i].style.display = 'none';
    }

    // -----------------------------
    // Diagonals per cell
    // -----------------------------
    if (diagonal) {
      let idx = 0;

      for (let cx = 0; cx < neededV; cx++) {
        for (let cy = 0; cy < neededH; cy++) {
          let line = diagonalPool[idx];
          if (!line) {
            line = document.createElementNS('http://www.w3.org/2000/svg', 'line');
            diagonalPool[idx] = line;
            gridGroup.appendChild(line);
          }

          const x = offsetX + cx * cellW;
          const y = offsetY + cy * cellH;

          line.setAttribute('x1', x);
          line.setAttribute('y1', y + cellH);
          line.setAttribute('x2', x + cellW);
          line.setAttribute('y2', y);
          line.setAttribute('stroke', stroke);
          line.setAttribute('stroke-width', 0.5);
          line.style.display = 'block';

          idx++;
        }
      }

      // hide unused
      for (let i = idx; i < diagonalPool.length; i++) {
        diagonalPool[i].style.display = 'none';
      }
    } else {
      // hide all diagonals
      for (let i = 0; i < diagonalPool.length; i++) {
        diagonalPool[i].style.display = 'none';
      }
    }
  };
};

function enableSnapIndicator(animator, measurer) {
  const snapV = document.getElementById('snap-v');
  const snapH = document.getElementById('snap-h');

  const SNAP_THRESHOLD = 8;

  svg.addEventListener('mousemove', (e) => {
    const cur = animator.getCurrent();
    const { offsetX, offsetY, cellW, cellH, snap } = cur;

    const mouseX = e.clientX;
    const mouseY = e.clientY;

    const isOutOfRange = mouseX < offsetX || mouseY < offsetY;

    // If snap disabled → hide and exit
    if (!snap || isOutOfRange) {
      snapV.setAttribute('visibility', 'hidden');
      snapH.setAttribute('visibility', 'hidden');
      return;
    }

    // nearest grid lines
    const col = Math.round((mouseX - offsetX) / cellW);
    const row = Math.round((mouseY - offsetY) / cellH);

    const snapX = offsetX + col * cellW;
    const snapY = offsetY + row * cellH;
    measurer.updateSnap(snapX, snapY);

    const dx = Math.abs(mouseX - snapX);
    const dy = Math.abs(mouseY - snapY);

    // vertical snap
    if (dx < SNAP_THRESHOLD) {
      snapV.setAttribute('x1', snapX);
      snapV.setAttribute('y1', 0);
      snapV.setAttribute('x2', snapX);
      snapV.setAttribute('y2', window.innerHeight);
      snapV.setAttribute('visibility', 'visible');
    } else {
      snapV.setAttribute('visibility', 'hidden');
    }

    // horizontal snap
    if (dy < SNAP_THRESHOLD) {
      snapH.setAttribute('x1', 0);
      snapH.setAttribute('y1', snapY);
      snapH.setAttribute('x2', window.innerWidth);
      snapH.setAttribute('y2', snapY);
      snapH.setAttribute('visibility', 'visible');
    } else {
      snapH.setAttribute('visibility', 'hidden');
    }
  });

  svg.addEventListener('mouseleave', () => {
    snapV.setAttribute('visibility', 'hidden');
    snapH.setAttribute('visibility', 'hidden');
  });
}

const createRulers = function () {
  const hTicks = [];
  const hLabels = [];
  const vTicks = [];
  const vLabels = [];

  const rulerH = document.getElementById('ruler-h');
  const rulerV = document.getElementById('ruler-v');

  return function drawRulers(settings) {
    const width = window.innerWidth;
    const height = window.innerHeight;

    const { offsetX, offsetY, cellW, cellH } = settings;
    const stroke = '#888';
    const fontSize = 10;

    // -----------------------------
    // Horizontal ruler (top)
    // -----------------------------
    let startCol = Math.floor((0 - offsetX) / cellW);
    let endCol = Math.ceil((width - offsetX) / cellW);

    let idx = 0;
    for (let col = startCol; col <= endCol; col++) {
      const x = offsetX + col * cellW;

      // Tick mark
      let tick = hTicks[idx];
      if (!tick) {
        tick = document.createElementNS('http://www.w3.org/2000/svg', 'line');
        hTicks[idx] = tick;
        rulerH.appendChild(tick);
      }

      tick.setAttribute('x1', x);
      tick.setAttribute('y1', 0);
      tick.setAttribute('x2', x);
      tick.setAttribute('y2', col % 5 === 0 ? 12 : 6);
      tick.setAttribute('stroke', stroke);
      tick.style.display = 'block';

      // Label every 5 cells
      if (col % 5 === 0) {
        let label = hLabels[col];
        if (!label) {
          label = document.createElementNS('http://www.w3.org/2000/svg', 'text');
          label.setAttribute('font-size', fontSize);
          label.setAttribute('fill', stroke);
          hLabels[col] = label;
          rulerH.appendChild(label);
        }

        label.setAttribute('x', x + 2);
        label.setAttribute('y', 22);
        label.textContent = col;
        label.style.display = 'block';
      }

      idx++;
    }

    // Hide unused horizontal ticks
    for (let i = idx; i < hTicks.length; i++) hTicks[i].style.display = 'none';
    // Hide unused labels
    Object.keys(hLabels).forEach((k) => {
      if (k < startCol || k > endCol) hLabels[k].style.display = 'none';
    });

    // -----------------------------
    // Vertical ruler (left)
    // -----------------------------
    let startRow = Math.floor((0 - offsetY) / cellH);
    let endRow = Math.ceil((height - offsetY) / cellH);

    idx = 0;
    for (let row = startRow; row <= endRow; row++) {
      const y = offsetY + row * cellH;

      // Tick mark
      let tick = vTicks[idx];
      if (!tick) {
        tick = document.createElementNS('http://www.w3.org/2000/svg', 'line');
        vTicks[idx] = tick;
        rulerV.appendChild(tick);
      }

      tick.setAttribute('x1', 0);
      tick.setAttribute('y1', y);
      tick.setAttribute('x2', row % 5 === 0 ? 12 : 6);
      tick.setAttribute('y2', y);
      tick.setAttribute('stroke', stroke);
      tick.style.display = 'block';

      // Label every 5 cells
      if (row % 5 === 0) {
        let label = vLabels[row];
        if (!label) {
          label = document.createElementNS('http://www.w3.org/2000/svg', 'text');
          label.setAttribute('font-size', fontSize);
          label.setAttribute('fill', stroke);
          vLabels[row] = label;
          rulerV.appendChild(label);
        }

        label.setAttribute('x', 14);
        label.setAttribute('y', y - 2);
        label.textContent = row;
        label.style.display = 'block';
      }

      idx++;
    }

    // Hide unused vertical ticks
    for (let i = idx; i < vTicks.length; i++) vTicks[i].style.display = 'none';
    // Hide unused labels
    Object.keys(vLabels).forEach((k) => {
      if (k < startRow || k > endRow) vLabels[k].style.display = 'none';
    });
  };
};

const createMeasure = function (settings, strokeWidth = 1.5, strokeColor = '#f00', textPadding = 40) {
  const measureH = [];
  const measureV = [];
  let _settings = settings;

  let hIndex = 0;
  let vIndex = 0;

  let lastSnapX = null;
  let lastSnapY = null;
  let dimH = null;
  let dimV = null;

  const updateSnap = (snapX, snapY) => {
    lastSnapX = snapX;
    lastSnapY = snapY;
  };

  const isLeftClick = (mouseEvent) => mouseEvent.button === 0;
  const isRightClick = (mouseEvent) => mouseEvent.button === 2;

  let clickPending = false;

  svg.addEventListener('mousedown', (e) => {
    clickPending = true;
    setTimeout(() => (clickPending = false), 150);

    const { snap } = _settings;
    if (!snap) return;
    if (isLeftClick(e)) addHorizontalMeasure();
    if (isRightClick(e)) addVerticalMeasure();
  });
  svg.addEventListener('contextmenu', (e) => e.preventDefault());

  window.addEventListener('focus', () => {
    const { snap } = _settings;

    setTimeout(() => {
      if (snap && clickPending) {
        ws.sendJSON({ type: 'clickedAndFocused' });
      }
      clickPending = false;
    }, 100);
  });

  const addVerticalMeasure = () => {
    if (lastSnapY === null) return;
    const { offsetY, cellH } = _settings;
    const raw = Math.round((lastSnapY - offsetY) / cellH);

    let line = measureH[hIndex]?.line;
    if (!line) {
      line = document.createElementNS('http://www.w3.org/2000/svg', 'line');
      line.setAttribute('stroke', strokeColor);
      line.setAttribute('stroke-width', strokeWidth);
      measureGroup.appendChild(line);
      measureH[hIndex] = { line, raw };
    }

    line.setAttribute('x1', 0);
    line.setAttribute('x2', window.innerWidth);
    line.setAttribute('y1', lastSnapY);
    line.setAttribute('y2', lastSnapY);
    line.style.display = 'block';

    hIndex = (hIndex + 1) % 2;
    updateDimensions(_settings);
  };

  const addHorizontalMeasure = function () {
    if (lastSnapX === null) return;
    const { offsetX, cellW } = _settings;
    const col = Math.round((lastSnapX - offsetX) / cellW);

    let line = measureV[vIndex]?.line;
    if (!line) {
      line = document.createElementNS('http://www.w3.org/2000/svg', 'line');
      line.setAttribute('stroke', strokeColor);
      line.setAttribute('stroke-width', strokeWidth);
      measureGroup.appendChild(line);
      measureV[vIndex] = { line, col };
    }

    line.setAttribute('y1', 0);
    line.setAttribute('y2', window.innerHeight);
    line.setAttribute('x1', lastSnapX);
    line.setAttribute('x2', lastSnapX);
    line.style.display = 'block';

    vIndex = (vIndex + 1) % 2;
    updateDimensions(_settings);
  };

  const drawHorizontalDimension = function (x1, x2, cellW, y) {
    if (!dimH) {
      dimH = {
        line: document.createElementNS('http://www.w3.org/2000/svg', 'line'),
        arrowL: document.createElementNS('http://www.w3.org/2000/svg', 'path'),
        arrowR: document.createElementNS('http://www.w3.org/2000/svg', 'path'),
        text: document.createElementNS('http://www.w3.org/2000/svg', 'text'),
      };

      Object.values(dimH).forEach((el) => {
        el.setAttribute('stroke', '#0f0');
        el.setAttribute('fill', '#0f0');
        el.setAttribute('stroke-width', 1.2);
        dimGroup.appendChild(el);
      });

      dimH.text.setAttribute('font-size', 12);
      dimH.text.setAttribute('fill', '#0f0');
    }

    // main line
    dimH.line.setAttribute('x1', x1);
    dimH.line.setAttribute('y1', y);
    dimH.line.setAttribute('x2', x2);
    dimH.line.setAttribute('y2', y);

    // arrow size
    const s = 6;

    // left arrow
    dimH.arrowL.setAttribute('d', `M ${x1} ${y} l ${s} ${-s / 2} l 0 ${s} Z`);

    // right arrow
    dimH.arrowR.setAttribute('d', `M ${x2} ${y} l ${-s} ${-s / 2} l 0 ${s} Z`);

    // text
    dimH.text.textContent = Math.trunc(Math.abs(x2 - x1) / cellW);
    dimH.text.setAttribute('x', (x1 + x2) / 2);
    dimH.text.setAttribute('y', y - 8);

    // show
    Object.values(dimH).forEach((el) => (el.style.display = 'block'));
  };

  const drawVerticalDimension = function (y1, y2, cellH, x) {
    if (!dimV) {
      dimV = {
        line: document.createElementNS('http://www.w3.org/2000/svg', 'line'),
        arrowT: document.createElementNS('http://www.w3.org/2000/svg', 'path'),
        arrowB: document.createElementNS('http://www.w3.org/2000/svg', 'path'),
        text: document.createElementNS('http://www.w3.org/2000/svg', 'text'),
      };

      Object.values(dimV).forEach((el) => {
        el.setAttribute('stroke', '#0f0');
        el.setAttribute('fill', '#0f0');
        el.setAttribute('stroke-width', 1.2);
        dimGroup.appendChild(el);
      });

      dimV.text.setAttribute('font-size', 12);
      dimV.text.setAttribute('fill', '#0f0');
    }

    // main line
    dimV.line.setAttribute('x1', x);
    dimV.line.setAttribute('y1', y1);
    dimV.line.setAttribute('x2', x);
    dimV.line.setAttribute('y2', y2);

    // arrow size
    const s = 6;

    // top arrow
    dimV.arrowT.setAttribute('d', `M ${x} ${y1} l ${-s / 2} ${s} l ${s} 0 Z`);

    // bottom arrow
    dimV.arrowB.setAttribute('d', `M ${x} ${y2} l ${-s / 2} ${-s} l ${s} 0 Z`);

    // text
    dimV.text.textContent = Math.trunc(Math.abs(y2 - y1) / cellH);
    dimV.text.setAttribute('x', x + 8);
    dimV.text.setAttribute('y', (y1 + y2) / 2);

    // show
    Object.values(dimV).forEach((el) => (el.style.display = 'block'));
  };

  function updateDimensions(settings) {
    const { cellH, cellW } = settings;
    if (measureV.length >= 2 && measureV[0] && measureV[1]) {
      const x1 = parseFloat(measureV[0].line.getAttribute('x1'));
      const x2 = parseFloat(measureV[1].line.getAttribute('x1'));
      drawHorizontalDimension(x1, x2, cellW, textPadding);
    } else if (dimH) {
      Object.values(dimH).forEach((el) => (el.style.display = 'none'));
    }

    if (measureH.length >= 2 && measureH[0] && measureH[1]) {
      const y1 = parseFloat(measureH[0].line.getAttribute('y1'));
      const y2 = parseFloat(measureH[1].line.getAttribute('y1'));
      drawVerticalDimension(y1, y2, cellH, textPadding);
    } else if (dimV) {
      Object.values(dimV).forEach((el) => (el.style.display = 'none'));
    }
  }

  const clearMeasurements = function () {
    measureH.forEach((l) => l.remove());
    measureV.forEach((l) => l.remove());
    measureH.length = 0;
    measureV.length = 0;
    if (dimH) {
      Object.values(dimH).forEach((el) => el.remove());
      dimH = null;
    }
    if (dimV) {
      Object.values(dimV).forEach((el) => el.remove());
      dimV = null;
    }
  };

  const update = function (settings) {
    _settings = settings;
    const { offsetX, offsetY, cellH, cellW } = settings;

    measureH.forEach((measureData) => {
      if (measureData && measureData.line && measureData.line.style.display !== 'none') {
        const { row } = measureData;
        const newY = offsetY + row * cellH;
        measureData.line.setAttribute('y1', newY);
        measureData.line.setAttribute('y2', newY);
      }
    });

    measureV.forEach((measureData) => {
      if (measureData && measureData.line && measureData.line.style.display !== 'none') {
        const { col } = measureData;
        const newX = offsetX + col * cellW;
        measureData.line.setAttribute('x1', newX);
        measureData.line.setAttribute('x2', newX);
      }
    });
    updateDimensions(settings);
  };

  return {
    updateSnap,
    addHorizontalMeasure,
    addVerticalMeasure,
    clearMeasurements,
    update,
  };
};

const drawGrid = createGrid();
const drawRulers = createRulers();
const measurer = createMeasure(defaultSettings, 0.5);

const animator = createGridAnimator((settings) => {
  drawGrid(settings);
  drawRulers(settings);
  measurer.update(settings);
}, defaultSettings);

animator.setDuration(50);
enableSnapIndicator(animator, measurer);
drawGrid(defaultSettings);

ws.onMessage((msg) => {
  console.log('on message', msg.type);
  if (msg.type === 'update' && msg.settings) {
    animator.push(msg.settings);
  }
});

ws.onStatusChange(function (status) {
  console.log('WS status:', status);
});

window.addEventListener('resize', () => {
  animator.push(animator.getCurrent());
});

window.addEventListener('keydown', (e) => {
  const { snap } = animator.getCurrent();
  if (!snap) return;
  if (e.key === 'x' || e.key === 'X') measurer.addHorizontalMeasure();
  if (e.key === 'y' || e.key === 'Y') measurer.addVerticalMeasure();
  if (e.key === 'd' || e.key === 'D' || e.key === 'Delete') measurer.clearMeasurements();
});
