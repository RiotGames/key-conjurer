import { defineConfig } from 'vite'
import Markdown, { Mode } from 'vite-plugin-markdown';
import react from '@vitejs/plugin-react';

export default defineConfig({
    plugins: [
        Markdown({ mode: [Mode.HTML] }),
        react()
    ],
    test: {
        environment: 'jsdom',
    }
});
