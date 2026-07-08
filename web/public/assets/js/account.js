import { apiCall, logout, initAuth } from './api.js';
import { setupPage, showToast, confirmAction } from './ui.js';

function validatePassword(password) {
    if (password.length < 8) {
        return 'Password must be at least 8 characters.';
    }
    return null;
}

async function changePassword(e) {
    e.preventDefault();
    const oldPassword = document.getElementById('oldPassword').value;
    const newPassword = document.getElementById('newPassword').value;

    if (!oldPassword || !newPassword) return;

    const passwordError = validatePassword(newPassword);
    if (passwordError) {
        showToast(passwordError, true);
        return;
    }

    try {
        const res = await apiCall('/users/me/password', {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ old_password: oldPassword, new_password: newPassword })
        });

        if (!res.ok) {
            const data = await res.json();
            throw new Error(data.error || 'Failed to change password');
        }
        showToast('Password changed successfully!');
        document.getElementById('changePasswordForm').reset();
    } catch (err) {
        showToast(err.message, true);
    }
}

async function toggleAccountStatus() {
    const badge = document.getElementById('accountStatusBadge');
    const isActive = badge && badge.textContent === 'Active';
    const newStatus = isActive ? 'inactive' : 'active';

    try {
        const res = await apiCall('/users/me/status', {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ status: newStatus })
        });

        if (!res.ok) throw new Error('Failed to update account status');

        if (badge) {
            badge.textContent = isActive ? 'Inactive' : 'Active';
            badge.className = isActive ? 'status-badge inactive' : 'status-badge active';
        }
        const btn = document.getElementById('toggleAccountStatusBtn');
        if (btn) btn.textContent = isActive ? 'Activate Account' : 'Deactivate Account';

        showToast(`Account is now ${newStatus}`);
    } catch (err) {
        showToast(err.message, true);
    }
}

async function deleteAccount() {
    confirmAction(
        'Delete Account',
        'WARNING: Are you absolutely sure you want to permanently delete your account? This cannot be undone.',
        async () => {
            try {
                const res = await apiCall('/users/me', { method: 'DELETE' });
                if (!res.ok) throw new Error('Failed to delete account');
                showToast('Account deleted successfully.');
                logout();
            } catch (err) {
                showToast(err.message, true);
            }
        }
    );
}

document.addEventListener('DOMContentLoaded', async () => {
    if (!setupPage()) return;
    await initAuth();

    const emailField = document.getElementById('accountEmail');
    const storedEmail = localStorage.getItem('user_email');
    if (emailField && storedEmail) {
        emailField.textContent = storedEmail;
    }

    const changePasswordForm = document.getElementById('changePasswordForm');
    if (changePasswordForm) {
        changePasswordForm.addEventListener('submit', (e) => changePassword(e));
    }

    const toggleAccountBtn = document.getElementById('toggleAccountStatusBtn');
    if (toggleAccountBtn) {
        toggleAccountBtn.addEventListener('click', () => toggleAccountStatus());
    }

    const deleteAccountBtn = document.getElementById('deleteAccountBtn');
    if (deleteAccountBtn) {
        deleteAccountBtn.addEventListener('click', () => deleteAccount());
    }
});
