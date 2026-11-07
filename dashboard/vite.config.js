import { defineConfig } from "vite";

export default defineConfig({
  base: "./",
  build: {
    outDir: "../arsy/console/dist",
    emptyOutDir: true,
  },
});
