import {fileURLToPath, URL} from 'node:url'
import {defineConfig} from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig(({command, mode, ssrBuild}) => {
    return {
        plugins: [vue()],
        resolve: {
            alias: {
                '@': fileURLToPath(new URL('./src', import.meta.url))
            }
        },
        server: {
            proxy: {
                // List them individually so Vite doesn't proxy itself!
                '/session': 'http://localhost:3000',
                '/users': 'http://localhost:3000',
                '/conversations': 'http://localhost:3000',
            },
        },
        define: {
            "__API_URL__": JSON.stringify(""), 
        },
    };
})