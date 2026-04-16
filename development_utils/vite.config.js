import path from 'path';

const PORT = 8182;

export default {
  root: './src/webpage',
  server: {
    port: PORT,
    open: false,
    fs: {
      strict: true,
    },
  },
  resolve: {
    alias: {
      '/static': path.resolve(import.meta.dirname, 'src/webpage'),
    },
  },
  plugins: [
    {
      name: 'ws-url-replacer-plugin',
      enforce: 'pre',
      transform(code, id) {
        const filesRegex = /(grid|navbar)\.js$/;

        if (filesRegex.test(id)) {
          const WS_PORT = process.env.PORT || 8181;
          const replacedCode = code.replace(/ws:\/\/\$\{location\.host\}/g, `ws://localhost:${WS_PORT}`);
          console.log('[ws-url-replacer-plugin] Replaced code in:', id);

          return {
            code: replacedCode,
            map: null,
          };
        }
      },
    },
  ],
};
