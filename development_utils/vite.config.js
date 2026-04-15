import path from "path";

const PORT = 8182;

export default {
  root: "./src/webpage",
  server: {
    port: PORT,
    open: false,
    fs: {
      strict: true,
    },
  },
  resolve: {
    alias: {
      "/static": path.resolve(import.meta.dirname, "src/webpage"),
    },
  },

};
