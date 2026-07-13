import { CONFIG } from './config.js';
import { showToast } from './ui.js';

let csrfToken = '';

export function getIsLoggedIn() {
    return localStorage.getItem('is_logged_in') === 'true';
}

export function setLoggedIn(status, email) {
    if (status) {
        localStorage.setItem('is_logged_in', 'true');
        if (email) localStorage.setItem('user_email', email);
    } else {
        localStorage.removeItem('is_logged_in');
        localStorage.removeItem('user_email');
    }
}

export async function initAuth() {
    if (getIsLoggedIn()) {
        try {
            const res = await fetch(`${CONFIG.API_URL}/csrf-token`, { credentials: 'include' });
            if (res.ok) {
                csrfToken = res.headers.get('X-CSRF-Token');
            } else {
                setLoggedIn(false);
            }
        } catch (err) {
            console.error("Failed to initialize auth:", err);
        }
    }
}

export async function apiCall(endpoint, options = {}) {
    options.credentials = 'include';
    if (!options.headers) options.headers = {};

    if (csrfToken && options.method && options.method !== 'GET' && options.method !== 'HEAD') {
        options.headers['X-CSRF-Token'] = csrfToken;
    }

    const res = await fetch(`${CONFIG.API_URL}${endpoint}`, options);
    
    if (res.status === 401 || res.status === 403) {
        if (endpoint !== '/users/login' && endpoint !== '/users/register') {
            showToast('Session expired. Please log in again.', true);
            await logout(false);
        }
    }
    return res;
}

export async function logout(callApi = false) {
    if (callApi) {
        try {
            await apiCall('/users/logout', { method: 'POST' });
        } catch (e) {
            console.error("Logout API failed", e);
        }
    }
    setLoggedIn(false);
    csrfToken = '';
    window.location.href = '/auth/login';
}
