export const createWSClient = function (url) {
  let socket = null;
  let reconnectTimer = null;
  let reconnectDelay = 500;
  const maxDelay = 5000;
  let isManualClose = false;

  let lastSentSettings = null;
  let onMessageHandler = () => {};
  let onStatusHandler = () => {};

  const parseMessage = async (data) => (data instanceof Blob ? JSON.parse(await data.text()) : JSON.parse(data));

  const scheduleReconnect = () => {
    clearTimeout(reconnectTimer);

    reconnectTimer = setTimeout(() => {
      reconnectDelay = Math.min(reconnectDelay * 1.5, maxDelay);
      connect();
    }, reconnectDelay);
  };

  const connect = () => {
    socket = new WebSocket(url);

    onStatusHandler('connecting');

    socket.onopen = () => {
      onStatusHandler('connected');
      reconnectDelay = 500;

      if (lastSentSettings) {
        socket.send(JSON.stringify(lastSentSettings));
      }
    };

    socket.onclose = () => {
      onStatusHandler('disconnected');
      if (!isManualClose) scheduleReconnect();
    };

    socket.onerror = () => {
      onStatusHandler('error');
      socket.close();
    };

    socket.onmessage = (event) => {
      parseMessage(event.data).then(onMessageHandler);
    };
  };

  const sendJSON = (obj) => {
    lastSentSettings = obj;
    if (socket.readyState === WebSocket.OPEN) {
      socket.send(JSON.stringify(obj));
    }
  };

  const close = () => {
    isManualClose = true;
    socket.close();
  };

  connect();

  return {
    sendJSON,
    close,
    onMessage: (fn) => (onMessageHandler = fn),
    onStatusChange: (fn) => (onStatusHandler = fn),
  };
};
