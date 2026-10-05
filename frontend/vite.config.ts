import react, { reactCompilerPreset } from '@vitejs/plugin-react'
import babel from '@rolldown/plugin-babel'

import { defineConfig } from 'vite'
import { resolve } from 'path'

// https://vite.dev/config/
export default defineConfig({
  build: {
    rolldownOptions: {
      output: {
        entryFileNames: "assets/[hash].js",
        chunkFileNames: "assets/[hash].js",
        assetFileNames: "assets/[hash].[ext]"
      }
    }
  },
  resolve: {
    alias: {
      "@": resolve(import.meta.dirname, "./src")
    }
  },
  plugins: [
    react(),
    babel({
      presets: [reactCompilerPreset()]
    })
  ]
})
