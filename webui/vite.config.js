import { fileURLToPath, URL } from 'node:url'

// Wrap dynamic import for Node 20.18 compatibility
const vuePlugin = async () => {
  const vue = await import('@vitejs/plugin-vue')
  return vue.default()
}

/**
 * https://vitejs.dev/config/
 */
export default async ({ command, mode, ssrBuild }) => {
  const ret = {
    plugins: [
      await vuePlugin() // dynamic import ensures Node 20.18 can load ESM
    ],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url))
      }
    },
  }

  ret.define = {
    // Do not modify this constant, it is used in the evaluation.
    "__API_URL__": JSON.stringify("http://localhost:3000"),
  }

  return ret
}
