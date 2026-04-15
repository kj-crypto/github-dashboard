import { createGridAnimator } from './utils.js';
import { createWSClient } from './wsClient.js';

const ws = createWSClient(`ws://${location.host}`);
const svg = document.getElementById('grid');
const gridGroup = document.getElementById('grid-lines');

const measureLine = document.getElementById('measure-line');
const measureText = document.getElementById('measure-text');

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

function enableSnapIndicator(animator) {
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

    const dx = Math.abs(mouseX - snapX);
    const dy = Math.abs(mouseY - snapY);

    // vertical snap
    if (dx < SNAP_THRESHOLD) {
      snapV.setAttribute('x1', snapX);
      snapV.setAttribute('y1', offsetY);
      snapV.setAttribute('x2', snapX);
      snapV.setAttribute('y2', window.innerHeight);
      snapV.setAttribute('visibility', 'visible');
    } else {
      snapV.setAttribute('visibility', 'hidden');
    }

    // horizontal snap
    if (dy < SNAP_THRESHOLD) {
      snapH.setAttribute('x1', offsetX);
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

const drawGrid = createGrid();
const drawRulers = createRulers();

const animator = createGridAnimator((settings) => {
  drawGrid(settings);
  drawRulers(settings);
}, defaultSettings);

animator.setDuration(20);
enableSnapIndicator(animator);

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

let measuring = false;
let startX = 0;
let startY = 0;

const pixelToGrid = function (x, y, settings) {
  const { offsetX, offsetY, cellW, cellH } = settings;
  return {
    col: (x - offsetX) / cellW,
    row: (y - offsetY) / cellH,
  };
};

svg.addEventListener('mousedown', (e) => {
  measuring = true;
  startX = e.clientX;
  startY = e.clientY;

  measureLine.setAttribute('x1', startX);
  measureLine.setAttribute('y1', startY);
  measureLine.setAttribute('x2', startX);
  measureLine.setAttribute('y2', startY);
  measureLine.setAttribute('visibility', 'visible');

  measureText.setAttribute('visibility', 'visible');
});

svg.addEventListener('mousemove', (e) => {
  if (!measuring) return;

  const x = e.clientX;
  const y = e.clientY;

  measureLine.setAttribute('x2', x);
  measureLine.setAttribute('y2', y);

  // Convert to grid units
  const cur = animator.getCurrent();
  const startGrid = pixelToGrid(startX, startY, cur);
  const endGrid = pixelToGrid(x, y, cur);

  const dCol = endGrid.col - startGrid.col;
  const dRow = endGrid.row - startGrid.row;
  const distCells = Math.sqrt(dCol * dCol + dRow * dRow);

  measureText.textContent = `ΔX: ${dCol.toFixed(2)}  ΔY: ${dRow.toFixed(2)}  Dist: ${distCells.toFixed(2)} cells`;

  measureText.setAttribute('x', x + 10);
  measureText.setAttribute('y', y - 10);
});

svg.addEventListener('mouseup', () => {
  measuring = false;
  measureLine.setAttribute('visibility', 'hidden');
  measureText.setAttribute('visibility', 'hidden');
});
