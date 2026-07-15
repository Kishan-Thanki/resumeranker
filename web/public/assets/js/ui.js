import { getIsLoggedIn, logout } from './api.js';

let toastContainer = null;

export function initGlobalComponents() {
    if (!document.getElementById('toastContainer')) {
        toastContainer = document.createElement('div');
        toastContainer.id = 'toastContainer';
        document.body.appendChild(toastContainer);
    } else {
        toastContainer = document.getElementById('toastContainer');
    }

    if (!document.getElementById('globalConfirmModal')) {
        const modal = document.createElement('div');
        modal.id = 'globalConfirmModal';
        modal.className = 'modal hidden';
        modal.setAttribute('role', 'dialog');
        modal.setAttribute('aria-modal', 'true');

        const content = document.createElement('div');
        content.className = 'modal-content card';

        const title = document.createElement('h3');
        title.id = 'confirmModalTitle';
        title.textContent = 'Are you sure?';

        const message = document.createElement('p');
        message.id = 'confirmModalMessage';
        message.className = 'mt-4 text-muted';
        message.textContent = 'This action cannot be undone.';

        const actions = document.createElement('div');
        actions.className = 'modal-actions';

        const btnCancel = document.createElement('button');
        btnCancel.className = 'btn btn-secondary';
        btnCancel.id = 'btnConfirmCancel';
        btnCancel.textContent = 'Cancel';

        const btnAccept = document.createElement('button');
        btnAccept.className = 'btn btn-danger';
        btnAccept.id = 'btnConfirmAccept';
        btnAccept.textContent = 'Confirm';

        actions.appendChild(btnCancel);
        actions.appendChild(btnAccept);
        content.appendChild(title);
        content.appendChild(message);
        content.appendChild(actions);
        modal.appendChild(content);
        document.body.appendChild(modal);
    }
}

export function showToast(message, isError = false) {
    if (!toastContainer) return;
    
    const toast = document.createElement('div');
    toast.className = 'toast';

    const icon = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
    icon.setAttribute('width', '18');
    icon.setAttribute('height', '18');
    icon.setAttribute('viewBox', '0 0 24 24');
    icon.setAttribute('fill', 'none');
    icon.setAttribute('stroke', 'currentColor');
    icon.setAttribute('stroke-width', '2');

    if (isError) {
        const circle = document.createElementNS('http://www.w3.org/2000/svg', 'circle');
        circle.setAttribute('cx', '12');
        circle.setAttribute('cy', '12');
        circle.setAttribute('r', '10');
        const line1 = document.createElementNS('http://www.w3.org/2000/svg', 'line');
        line1.setAttribute('x1', '15'); line1.setAttribute('y1', '9');
        line1.setAttribute('x2', '9'); line1.setAttribute('y2', '15');
        const line2 = document.createElementNS('http://www.w3.org/2000/svg', 'line');
        line2.setAttribute('x1', '9'); line2.setAttribute('y1', '9');
        line2.setAttribute('x2', '15'); line2.setAttribute('y2', '15');
        icon.appendChild(circle);
        icon.appendChild(line1);
        icon.appendChild(line2);
    } else {
        const path = document.createElementNS('http://www.w3.org/2000/svg', 'path');
        path.setAttribute('d', 'M22 11.08V12a10 10 0 1 1-5.93-9.14');
        const polyline = document.createElementNS('http://www.w3.org/2000/svg', 'polyline');
        polyline.setAttribute('points', '22 4 12 14.01 9 11.01');
        icon.appendChild(path);
        icon.appendChild(polyline);
    }

    const span = document.createElement('span');
    span.textContent = message;

    toast.appendChild(icon);
    toast.appendChild(span);
    toastContainer.appendChild(toast);

    setTimeout(() => {
        if (toast.parentElement) toast.remove();
    }, 3500);
}

export function confirmAction(title, message, onConfirm) {
    const modal = document.getElementById('globalConfirmModal');
    if (!modal) return;
    
    document.getElementById('confirmModalTitle').textContent = title;
    document.getElementById('confirmModalMessage').textContent = message;

    const btnCancel = document.getElementById('btnConfirmCancel');
    const btnAccept = document.getElementById('btnConfirmAccept');

    const newBtnCancel = btnCancel.cloneNode(true);
    const newBtnAccept = btnAccept.cloneNode(true);
    btnCancel.parentNode.replaceChild(newBtnCancel, btnCancel);
    btnAccept.parentNode.replaceChild(newBtnAccept, btnAccept);

    newBtnCancel.addEventListener('click', () => {
        modal.classList.add('hidden');
    });

    newBtnAccept.addEventListener('click', () => {
        modal.classList.add('hidden');
        onConfirm();
    });

    modal.classList.remove('hidden');
}

export function setLoading(btn, isLoading) {
    if (!btn) return;
    const text = btn.querySelector('.btn-text');
    const loader = btn.querySelector('.loader');

    if (isLoading) {
        btn.disabled = true;
        if (text) text.classList.add('hidden');
        if (loader) loader.classList.remove('hidden');
    } else {
        btn.disabled = false;
        if (text) text.classList.remove('hidden');
        if (loader) loader.classList.add('hidden');
    }
}

export function initTheme() {
    const toggleBtn = document.getElementById('themeToggleBtn');
    if (toggleBtn) {
        toggleBtn.addEventListener('click', () => {
            const root = document.documentElement;
            const current = root.getAttribute('data-theme');
            const next = current === 'dark' ? 'light' : 'dark';
            root.setAttribute('data-theme', next);
            localStorage.setItem('data-theme', next);
        });
    }

    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    mediaQuery.addEventListener('change', (e) => {
        if (!localStorage.getItem('data-theme')) {
            document.documentElement.setAttribute('data-theme', e.matches ? 'dark' : 'light');
        }
    });
}

export function initScrollListener() {
    const navbar = document.querySelector('.navbar');
    if (!navbar) return;
    
    const onScroll = () => {
        if (window.scrollY > 10) {
            navbar.classList.add('scrolled');
        } else {
            navbar.classList.remove('scrolled');
        }
    };
    
    window.addEventListener('scroll', onScroll, { passive: true });
    onScroll();
}

export function setActiveNavLink() {
    const path = window.location.pathname;
    const links = document.querySelectorAll('.nav-links a');
    
    links.forEach(link => {
        const href = link.getAttribute('href');
        if (href === '/' && path === '/') {
            link.classList.add('active');
        } else if (href !== '/' && path.startsWith(href)) {
            link.classList.add('active');
        } else {
            link.classList.remove('active');
        }
    });
}

export function bindGlobalEvents() {
    const navLogout = document.getElementById('navLogout');
    if (navLogout) {
        navLogout.addEventListener('click', (e) => {
            e.preventDefault();
            logout(true);
        });
    }

    const hamburger = document.getElementById('navHamburger');
    const navLinks = document.querySelector('.nav-links');
    if (hamburger && navLinks) {
        hamburger.addEventListener('click', () => {
            hamburger.classList.toggle('active');
            navLinks.classList.toggle('open');
        });

        navLinks.querySelectorAll('a').forEach(link => {
            link.addEventListener('click', () => {
                hamburger.classList.remove('active');
                navLinks.classList.remove('open');
            });
        });
    }
}

export function initNavigation() {
    const isLoggedIn = getIsLoggedIn();
    const navLogout = document.getElementById('navLogout');
    const navLoginBtn = document.getElementById('navLoginBtn');
    const navDashBtn = document.getElementById('navDashBtn');

    if (isLoggedIn) {
        if (navLogout) navLogout.classList.remove('hidden');
        if (navDashBtn) navDashBtn.classList.remove('hidden');
        if (navLoginBtn) navLoginBtn.classList.add('hidden');
    } else {
        if (navLogout) navLogout.classList.add('hidden');
        if (navDashBtn) navDashBtn.classList.add('hidden');
        if (navLoginBtn) navLoginBtn.classList.remove('hidden');
    }
}

export function checkProtectedRoutes() {
    const path = window.location.pathname;
    const isLoggedIn = getIsLoggedIn();
    const protectedRoutes = ['/dashboard', '/dashboard/guide', '/dashboard/account'];
    const isProtected = protectedRoutes.some(route => path.includes(route));

    if (isProtected && !isLoggedIn) {
        window.location.href = '/auth/login';
        return false;
    }

    const isPublicHome = path === '/' || path === '' || path.includes('index');
    if (isLoggedIn && (path.includes('login') || path.includes('register'))) {
        window.location.href = '/dashboard';
        return false;
    }
    return true;
}

export function setupPage() {
    initGlobalComponents();
    initTheme();
    initNavigation();
    initScrollListener();
    setActiveNavLink();
    bindGlobalEvents();
    return checkProtectedRoutes();
}
