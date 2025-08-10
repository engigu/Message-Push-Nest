/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_RUN_MODE: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}

// 全局构建时间变量
declare const __BUILD_TIME__: string