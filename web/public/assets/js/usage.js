import { apiCall, initAuth } from './api.js';
import { setupPage } from './ui.js';

function escapeHTML(str) {
    if (!str) return '';
    const div = document.createElement('div');
    div.textContent = str;
    return div.innerHTML;
}

document.addEventListener('DOMContentLoaded', async () => {
    if (!setupPage()) return;
    await initAuth();

    const loading = document.getElementById('usageLoading');
    const container = document.getElementById('usageContainer');

    try {
        const response = await apiCall('/keys');
        if (!response.ok) {
            throw new Error('Failed to load API keys');
        }

        const keys = await response.json();
        
        // Fetch real-time stats for each key
        const statsPromises = keys.map(key => apiCall(`/keys/${key.id}/stats`).then(res => res.json()));
        const allStats = await Promise.all(statsPromises);
        
        loading.style.display = 'none';
        container.style.display = 'block';

        if (!keys || keys.length === 0) {
            container.innerHTML = `
                <div class="card card-narrow">
                    <p class="text-muted text-center" style="margin: 2rem 0;">
                        You have not generated any API keys yet.<br>
                        <a href="/dashboard" style="color: var(--primary-color);">Go to API Keys</a> to generate one.
                    </p>
                </div>
            `;
            return;
        }

        let html = '';
        keys.forEach((key, index) => {
            const stats = allStats[index];
            const tokensUsed = key.tokens_used || 0;
            const tokenQuota = key.token_quota || 0;
            
            // Calculate percentage safely
            let percentage = 0;
            if (tokenQuota > 0) {
                percentage = Math.min(100, Math.round((tokensUsed / tokenQuota) * 100));
            }
            
            // Determine progress bar color based on usage
            let progressClass = 'bg-primary';
            if (percentage >= 90) progressClass = 'bg-danger';
            else if (percentage >= 75) progressClass = 'bg-warning';

            html += `
                <div class="usage-card-row">
                    <div class="usage-col-identity">
                        <div class="usage-title-row">
                            <span class="status-dot ${key.status === 'active' ? 'status-active' : 'status-inactive'}"></span>
                            <h3>${escapeHTML(key.name)}</h3>
                        </div>
                        <div class="usage-prefix-row">
                            <code class="mono-badge">${escapeHTML(key.key_prefix)}••••••</code>
                        </div>
                    </div>
                    
                    <div class="usage-col-metrics">
                        <div class="metric-box">
                            <span class="metric-label">RPM</span>
                            <span class="metric-value">${stats.rpm_used} <span class="metric-limit">/ ${stats.rpm_limit}</span></span>
                        </div>
                        <div class="metric-box">
                            <span class="metric-label">RPD</span>
                            <span class="metric-value">${stats.rpd_used} <span class="metric-limit">/ ${stats.rpd_limit}</span></span>
                        </div>
                    </div>

                    <div class="usage-col-volume" style="display: flex; align-items: center;">
                        <div class="progress-ring-container">
                            <svg class="progress-ring" width="72" height="72">
                                <circle class="progress-ring__circle-bg" stroke="var(--bg-hover)" stroke-width="5" fill="transparent" r="30" cx="36" cy="36"/>
                                <circle class="progress-ring__circle" stroke="var(--text-primary)" stroke-width="5" fill="transparent" r="30" cx="36" cy="36" stroke-dasharray="188.5" stroke-dashoffset="${188.5 - (percentage / 100) * 188.5}"/>
                            </svg>
                            <div class="progress-ring-text">${percentage}%</div>
                        </div>
                        <div class="volume-details">
                            <div class="volume-header">
                                <span class="metric-label">Lifetime Tokens</span>
                            </div>
                            <div class="volume-footer" style="text-align: left;">
                                <span class="metric-value">${tokensUsed.toLocaleString()} <span class="metric-limit">/ ${tokenQuota.toLocaleString()}</span></span>
                            </div>
                        </div>
                    </div>
                </div>
            `;
        });

        container.innerHTML = html;

    } catch (error) {
        console.error('Error loading usage:', error);
        loading.style.display = 'none';
        container.style.display = 'block';
        container.innerHTML = `
            <div class="alert alert-error">
                Failed to load usage data. Please try again later.
            </div>
        `;
    }
});
