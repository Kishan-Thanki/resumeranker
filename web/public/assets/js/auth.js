import { apiCall, setLoggedIn, initAuth } from './api.js';
import { setupPage, setLoading, showToast } from './ui.js';

function validatePassword(password) {
    if (password.length < 8) {
        return 'Password must be at least 8 characters.';
    }
    return null;
}

async function handleAuth(e, mode) {
    e.preventDefault();

    const errorEl = document.getElementById(`${mode}Error`);
    const submitBtn = document.getElementById(`${mode}SubmitBtn`);

    if (errorEl) errorEl.classList.add('hidden');
    setLoading(submitBtn, true);

    const email = document.getElementById(`${mode}Email`).value.trim();
    const password = document.getElementById(`${mode}Password`).value;

    if (mode === 'register') {
        const passwordError = validatePassword(password);
        if (passwordError) {
            if (errorEl) {
                errorEl.textContent = passwordError;
                errorEl.classList.remove('hidden');
            }
            setLoading(submitBtn, false);
            return;
        }
    }

    const endpoint = mode === 'login' ? '/users/login' : '/users/register';
    const body = mode === 'login'
        ? { email, password }
        : { email, password, role: 'customer', agreed_to_terms: document.getElementById('registerTerms')?.checked || false };

    try {
        const res = await apiCall(endpoint, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(body)
        });

        if (!res.ok) {
            let errMsg = 'Authentication failed';
            try {
                const data = await res.clone().json();
                errMsg = data.error || errMsg;
            } catch (parseErr) {
                errMsg = await res.text();
            }
            throw new Error(errMsg);
        }

        if (mode === 'register') {
            const form = document.getElementById('registerForm');
            const successEl = document.getElementById('registerSuccess');
            const emailSpan = document.getElementById('registeredEmail');
            const footer = document.getElementById('registerFooter');
            
            if (form && successEl) {
                form.classList.add('hidden');
                if (footer) footer.classList.add('hidden');
                if (emailSpan) emailSpan.textContent = email;
                successEl.classList.remove('hidden');
            } else {
                showToast('Registration successful! Please check your email to verify your account.', false);
                window.location.href = '/auth/login';
            }
        } else {
            setLoggedIn(true, email);
            window.location.href = '/dashboard';
        }
    } catch (err) {
        if (errorEl) {
            errorEl.textContent = err.message;
            errorEl.classList.remove('hidden');
        }
    } finally {
        setLoading(submitBtn, false);
    }
}

async function loginHelper(email, password) {
    const res = await apiCall('/users/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password })
    });

    if (!res.ok) throw new Error('Login failed after registration');

    setLoggedIn(true, email);
    window.location.href = '/dashboard';
}

document.addEventListener('DOMContentLoaded', async () => {
    if (!setupPage()) return;
    await initAuth();

    const loginForm = document.getElementById('loginForm');
    if (loginForm) {
        loginForm.addEventListener('submit', (e) => handleAuth(e, 'login'));
    }

    const registerForm = document.getElementById('registerForm');
    if (registerForm) {
        registerForm.addEventListener('submit', (e) => handleAuth(e, 'register'));
    }
});
