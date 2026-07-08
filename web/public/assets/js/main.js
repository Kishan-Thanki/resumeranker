import { setupPage } from './ui.js';
import { initAuth } from './api.js';

document.addEventListener('DOMContentLoaded', async () => {
    if (setupPage()) {
        await initAuth();
    }
});
