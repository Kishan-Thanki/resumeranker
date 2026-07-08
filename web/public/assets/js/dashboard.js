import { apiCall, logout, initAuth } from './api.js';
import { setupPage, showToast, confirmAction, setLoading } from './ui.js';

let keysTableBody;
let btnGenerateKey;
let noKeysMsg;
let newKeyAlert;
let newKeyValue;
let btnCopyKey;
let maxKeysMsg;
let keyModal;
let btnCancelKey;
let btnConfirmKey;
let keyNameInput;

async function loadKeys() {
    try {
        const res = await apiCall('/keys');
        if (res.status === 401) {
            logout();
            return;
        }
        const data = await res.json();
        renderKeys(data);
    } catch (err) {
        showToast('Failed to load API keys', true);
    }
}

function renderKeys(keys) {
    if (!keysTableBody) return;
    
    while (keysTableBody.firstChild) {
        keysTableBody.removeChild(keysTableBody.firstChild);
    }

    const hasKeys = keys && keys.length > 0;
    if (btnGenerateKey) btnGenerateKey.classList.toggle('hidden', hasKeys);
    if (maxKeysMsg) maxKeysMsg.classList.toggle('hidden', !hasKeys);

    if (!hasKeys) {
        if (noKeysMsg) noKeysMsg.classList.remove('hidden');
        const tableResp = document.querySelector('.table-responsive');
        if (tableResp) tableResp.classList.add('hidden');
        return;
    }

    if (noKeysMsg) noKeysMsg.classList.add('hidden');
    const tableResp = document.querySelector('.table-responsive');
    if (tableResp) tableResp.classList.remove('hidden');

    keys.forEach(key => {
        const tr = document.createElement('tr');

        const tdName = document.createElement('td');
        tdName.textContent = key.name;
        tr.appendChild(tdName);

        const tdPrefix = document.createElement('td');
        tdPrefix.className = 'mono';
        tdPrefix.textContent = key.key_prefix + '••••••';
        tr.appendChild(tdPrefix);

        const tdQuota = document.createElement('td');
        tdQuota.textContent = key.token_quota;
        tr.appendChild(tdQuota);

        const tdUsed = document.createElement('td');
        tdUsed.textContent = key.tokens_used;
        tr.appendChild(tdUsed);

        const tdExpires = document.createElement('td');
        tdExpires.textContent = key.expires_at ? new Date(key.expires_at).toLocaleDateString() : 'Never';
        tr.appendChild(tdExpires);

        const tdStatus = document.createElement('td');
        const badge = document.createElement('span');
        const isActive = key.status === 'active';
        badge.className = isActive ? 'status-badge active' : 'status-badge inactive';
        badge.textContent = isActive ? 'Active' : 'Inactive';
        tdStatus.appendChild(badge);
        tr.appendChild(tdStatus);

        const tdActions = document.createElement('td');
        tdActions.className = 'table-actions';

        const toggleBtn = document.createElement('button');
        toggleBtn.className = 'btn btn-secondary btn-sm';
        toggleBtn.textContent = isActive ? 'Deactivate' : 'Activate';
        toggleBtn.setAttribute('aria-label', isActive ? `Deactivate key ${key.name}` : `Activate key ${key.name}`);
        toggleBtn.addEventListener('click', () => {
            toggleKeyStatus(key.id, isActive ? 'inactive' : 'active');
        });

        const deleteBtn = document.createElement('button');
        deleteBtn.className = 'btn btn-danger btn-sm';
        deleteBtn.textContent = 'Delete';
        deleteBtn.setAttribute('aria-label', `Delete key ${key.name}`);
        deleteBtn.addEventListener('click', () => {
            revokeKey(key.id);
        });

        tdActions.appendChild(toggleBtn);
        tdActions.appendChild(deleteBtn);
        tr.appendChild(tdActions);

        keysTableBody.appendChild(tr);
    });
}

async function toggleKeyStatus(id, newStatus) {
    try {
        const res = await apiCall(`/keys/${id}/status`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ status: newStatus })
        });
        if (!res.ok) throw new Error('Failed to update status');
        loadKeys();
    } catch (err) {
        showToast(err.message, true);
    }
}

async function generateKey() {
    if (!keyNameInput) return;

    const name = keyNameInput.value.trim() || 'Default Key';

    setLoading(btnConfirmKey, true);
    if (newKeyAlert) newKeyAlert.classList.add('hidden');

    try {
        const res = await apiCall('/keys/generate', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name: name })
        });

        const data = await res.json();
        if (!res.ok) {
            showToast(data.error || 'Failed to generate key', true);
            return;
        }

        if (newKeyValue) newKeyValue.textContent = data.key;
        if (newKeyAlert) newKeyAlert.classList.remove('hidden');
        if (keyModal) keyModal.classList.add('hidden');

        showToast('API Key generated successfully');
        loadKeys();
    } catch (err) {
        showToast('Failed to generate key: ' + err.message, true);
    } finally {
        setLoading(btnConfirmKey, false);
    }
}

async function revokeKey(id) {
    confirmAction(
        'Revoke API Key',
        'Revoking this API key will instantly break any applications using it. Are you sure?',
        async () => {
            try {
                const res = await apiCall(`/keys/${id}`, { method: 'DELETE' });
                if (!res.ok) throw new Error('Failed to revoke');
                showToast('API Key revoked');
                loadKeys();
                if (newKeyAlert) newKeyAlert.classList.add('hidden');
            } catch (err) {
                showToast(err.message, true);
            }
        }
    );
}

document.addEventListener('DOMContentLoaded', async () => {
    if (!setupPage()) return;
    await initAuth();

    btnGenerateKey = document.getElementById('btnGenerateKey');
    keysTableBody = document.getElementById('keysTableBody');
    noKeysMsg = document.getElementById('noKeysMsg');
    newKeyAlert = document.getElementById('newKeyAlert');
    newKeyValue = document.getElementById('newKeyValue');
    btnCopyKey = document.getElementById('btnCopyKey');
    maxKeysMsg = document.getElementById('maxKeysMsg');
    keyModal = document.getElementById('keyModal');
    btnCancelKey = document.getElementById('btnCancelKey');
    btnConfirmKey = document.getElementById('btnConfirmKey');
    keyNameInput = document.getElementById('keyNameInput');

    if (btnGenerateKey) {
        btnGenerateKey.addEventListener('click', () => {
            if (keyModal) {
                keyModal.classList.remove('hidden');
                keyNameInput.value = '';
                keyNameInput.focus();
            }
        });
    }

    if (btnCancelKey) {
        btnCancelKey.addEventListener('click', () => {
            keyModal.classList.add('hidden');
        });
    }

    if (btnConfirmKey) {
        btnConfirmKey.addEventListener('click', () => generateKey());
    }

    if (btnCopyKey) {
        btnCopyKey.addEventListener('click', async () => {
            if (newKeyValue && newKeyValue.textContent) {
                try {
                    await navigator.clipboard.writeText(newKeyValue.textContent);
                    showToast('Copied to clipboard!');
                } catch (err) {
                    showToast('Failed to copy to clipboard', true);
                }
            }
        });
    }

    loadKeys();
});
