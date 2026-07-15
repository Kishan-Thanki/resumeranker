import { apiCall, logout, initAuth } from './api.js';
import { setupPage, showToast } from './ui.js';

document.addEventListener('DOMContentLoaded', async () => {
    setupPage();
    await initAuth();
    loadHistory();
});

async function loadHistory() {
    const historyContainer = document.getElementById('historyContainer');
    try {
        const res = await apiCall('/dashboard/analyze/history');
        if (res.status === 401) {
            logout();
            return;
        }
        if (!res.ok) throw new Error('Failed to load history');
        
        const data = await res.json();
        renderHistory(data);
    } catch (err) {
        showToast('Failed to load analysis history', true);
        if (historyContainer) {
            historyContainer.innerHTML = '<div class="alert alert-error" style="margin-top: 2rem;">Failed to load history. Please try again later.</div>';
        }
    }
}

function timeAgo(dateParam) {
    if (!dateParam) return null;
    const date = typeof dateParam === 'object' ? dateParam : new Date(dateParam);
    const today = new Date();
    const seconds = Math.round((today - date) / 1000);
    const minutes = Math.round(seconds / 60);
    const hours = Math.round(minutes / 60);
    const days = Math.round(hours / 24);

    if (seconds < 60) return 'Just now';
    if (minutes < 60) return `${minutes}m ago`;
    if (hours < 24) return `${hours}h ago`;
    if (days === 1) return 'Yesterday';
    if (days < 7) return `${days}d ago`;
    
    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
}

function renderHistory(items) {
    const historyContainer = document.getElementById('historyContainer');
    if (!historyContainer) return;
    
    if (!items || items.length === 0) {
        historyContainer.innerHTML = `
            <div class="history-empty-state">
                <svg class="history-empty-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1" stroke-linecap="round" stroke-linejoin="round">
                    <rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect>
                    <line x1="9" y1="3" x2="9" y2="21"></line>
                </svg>
                <h3>No analyses yet</h3>
                <p>You haven't run any resume analyses. Upload your first resume and job description to see the magic happen.</p>
                <a href="/dashboard/analyze" class="btn btn-primary">Analyze Your First Resume</a>
            </div>
        `;
        return;
    }
    
    let html = '<div class="history-grid">';
    
    items.forEach(item => {
        const reqId = item.request_id || item.id;
        const metadata = item.metadata || {};
        const jdName = metadata.jd_filename || item.jd_filename || 'Unknown JD';
        const resName = metadata.resume_filename || item.resume_filename || 'Unknown Resume';
        const relativeTime = timeAgo(item.created_at);
        const fullDate = new Date(item.created_at).toLocaleString();
        
        let statusBadge = '';
        if (item.status === 'completed') {
            statusBadge = `
                <div class="status-badge completed">
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"></polyline></svg>
                    Completed
                </div>
            `;
        } else if (item.status === 'failed') {
            statusBadge = `
                <div class="status-badge failed">
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>
                    Failed
                </div>
            `;
        } else {
            statusBadge = `
                <div class="status-badge processing">
                    <svg class="loader" style="width: 12px; height: 12px;" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"></svg>
                    Processing
                </div>
            `;
        }
        
        const docIcon = `<svg class="file-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path><polyline points="14 2 14 8 20 8"></polyline><line x1="16" y1="13" x2="8" y2="13"></line><line x1="16" y1="17" x2="8" y2="17"></line><polyline points="10 9 9 9 8 9"></polyline></svg>`;
        const briefIcon = `<svg class="file-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="7" width="20" height="14" rx="2" ry="2"></rect><path d="M16 21V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v16"></path></svg>`;

        html += `
            <div class="history-card">
                <div class="history-card-header">
                    <span class="date-text" title="${fullDate}">${relativeTime}</span>
                    ${statusBadge}
                </div>
                <div class="history-card-body">
                    <div class="file-info" title="${resName}">
                        ${docIcon}
                        <span class="file-name">${resName.replace('.pdf', '')}</span>
                    </div>
                    <div class="file-info" title="${jdName}">
                        ${briefIcon}
                        <span class="file-name">${jdName.replace('.pdf', '')}</span>
                    </div>
                </div>
                ${item.status === 'completed' ? `
                <div class="history-card-footer">
                    <a href="/dashboard/report?id=${reqId}&res=${encodeURIComponent(resName)}&jd=${encodeURIComponent(jdName)}" class="action-btn w-full">
                        View Report
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="5" y1="12" x2="19" y2="12"></line><polyline points="12 5 19 12 12 19"></polyline></svg>
                    </a>
                </div>
                ` : ''}
            </div>
        `;
    });
    
    html += '</div>';
    historyContainer.innerHTML = html;
}
