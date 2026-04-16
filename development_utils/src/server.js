import express from 'express';
import http from 'http';
import { WebSocketServer } from 'ws';
import path from 'path';

const PORT = process.env.PORT || 8181;
const app = express();

const isVite = process.env.VITE_DEV_SERVER === 'true';
if (!isVite) {
  app.use('/static', express.static(path.join(import.meta.dirname, 'webpage')));

  app.get('/', (req, res) => {
    res.sendFile(path.join(import.meta.dirname, 'webpage/index.html'));
  });

  app.get('/grid', (req, res) => {
    res.sendFile(path.join(import.meta.dirname, 'webpage/grid.html'));
  });
}

const server = http.createServer(app);
const wss = new WebSocketServer({ server });

wss.on('connection', (ws) => {
  console.log('Client connected');

  ws.on('message', (msg) => {
    for (const client of wss.clients) {
      if (client !== ws && client.readyState === 1) {
        client.send(msg);
      }
    }
  });

  ws.on('close', () => {
    console.log('Client disconnected');
  });
});

server.listen(PORT, () => {
  console.log(`Server running at http://localhost:${PORT}`);
});
