import { apiCall, logout, initAuth } from './api.js';
import { setupPage, showToast } from './ui.js';

document.addEventListener('DOMContentLoaded', async () => {
    setupPage();
    await initAuth();

    const urlParams = new URLSearchParams(window.location.search);
    const id = urlParams.get('id');
    const resName = urlParams.get('res') || 'Resume';
    const jdName = urlParams.get('jd') || 'Job Description';
    
    if (id) {
        function formatTitle(name) {
            if (!name) return '';
            return name.replace(/\.[^/.]+$/, "").replace(/[_-]/g, " ");
        }
        
        const subtitleEl = document.getElementById('reportSubtitle');
        if (subtitleEl) {
            subtitleEl.textContent = `${formatTitle(resName)} vs ${formatTitle(jdName)}`;
        }
        fetchResult(id);
    } else {
        document.getElementById('reportContent').innerHTML = '<div class="text-danger">Invalid report ID</div>';
    }
});

async function fetchResult(id) {
    try {
        const res = await apiCall(`/dashboard/analyze/${id}/result`);
        if (res.status === 401) {
            logout();
            return;
        }
        if (!res.ok) {
            const data = await res.json().catch(() => ({}));
            throw new Error(data.error || 'Failed to load report result');
        }
        
        const data = await res.json();
        
        // Parse the inner JSON if needed
        let resultJson = data.result;
        if (typeof resultJson === 'string') {
            try {
                resultJson = JSON.parse(resultJson);
            } catch (e) {
                console.error("Failed to parse result JSON", e);
                throw new Error("Invalid result format from server");
            }
        }
        
        renderReport(resultJson);
    } catch (err) {
        showToast(err.message, true);
        document.getElementById('reportContent').innerHTML = `<div class="text-danger">${err.message}</div>`;
    }
}

function renderReport(data) {
    const container = document.getElementById('reportContent');
    
    let score = data.score || data.match_score || 0;
    if (score === 0 && data.sections) {
        let total = 0, count = 0;
        data.sections.forEach(s => {
            if (s.score !== undefined) { total += s.score; count++; }
        });
        score = count > 0 ? Math.round(total / count) : 0;
    }
    
    let html = `
        <div class="report-hero card mb-4" style="text-align: center; padding: 3rem 2rem;">
            <div style="font-size: 1.5rem; font-weight: 600; margin-bottom: 1rem;">Overall Match Score</div>
            <div class="score-circle" style="width: 120px; height: 120px; border-radius: 50%; background: conic-gradient(var(--success) ${score}%, var(--bg-surface-elevated) 0); display: flex; align-items: center; justify-content: center; margin: 0 auto; position: relative;">
                <div style="width: 100px; height: 100px; border-radius: 50%; background: var(--bg-surface); display: flex; align-items: center; justify-content: center; font-size: 2.5rem; font-weight: 700;">${score}</div>
            </div>
            <div class="mt-4 text-body" style="max-width: 800px; margin: 2rem auto 0; text-align: left; line-height: 1.6;">
                ${data.completeAnalysis ? data.completeAnalysis : 'No executive summary provided.'}
            </div>
        </div>
    `;
    
    if (data.sections && data.sections.length > 0) {
        html += `<div class="report-tabs" style="display: flex; gap: 1rem; margin-bottom: 1rem; border-bottom: 1px solid var(--border-color); overflow-x: auto;">`;
        data.sections.forEach((sec, idx) => {
            const active = idx === 0 ? 'border-bottom: 2px solid var(--accent); font-weight: 600;' : 'border-bottom: 2px solid transparent; color: var(--text-muted); cursor: pointer;';
            html += `<div class="report-tab" data-target="tab-${sec.id}" style="padding: 0.75rem 1.5rem; white-space: nowrap; transition: all 0.2s; ${active}" onclick="switchTab('${sec.id}')">${sec.label} (${sec.score}%)</div>`;
        });
        html += `</div>`;
        
        html += `<div class="report-tab-contents">`;
        data.sections.forEach((sec, idx) => {
            const display = idx === 0 ? 'block' : 'none';
            html += `<div class="report-tab-content" id="tab-${sec.id}" style="display: ${display};">`;
            
            if (data.sectionsAnalysis && data.sectionsAnalysis[sec.id]) {
                html += `<div class="card mb-4" style="background: var(--bg-surface-elevated);"><p class="text-body m-0">${data.sectionsAnalysis[sec.id]}</p></div>`;
            }
            
            if (sec.requirements && sec.requirements.length > 0) {
                html += `<div class="requirements-grid" style="display: grid; gap: 1.5rem;">`;
                sec.requirements.forEach(req => {
                    let icon = '', badge = '';
                    if (req.matched) {
                        if (req.matchStrength === 'strong') {
                            icon = `<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="var(--success)" stroke-width="2"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path><polyline points="22 4 12 14.01 9 11.01"></polyline></svg>`;
                            badge = `<span style="background: var(--success-bg); color: var(--success); padding: 0.25rem 0.5rem; border-radius: 4px; font-size: 0.75rem; font-weight: 600;">Strong Match</span>`;
                        } else {
                            icon = `<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#eab308" stroke-width="2"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path><line x1="12" y1="9" x2="12" y2="13"></line><line x1="12" y1="17" x2="12.01" y2="17"></line></svg>`;
                            badge = `<span style="background: rgba(234, 179, 8, 0.1); color: #eab308; padding: 0.25rem 0.5rem; border-radius: 4px; font-size: 0.75rem; font-weight: 600;">Partial Match</span>`;
                        }
                    } else {
                        icon = `<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="var(--error)" stroke-width="2"><circle cx="12" cy="12" r="10"></circle><line x1="15" y1="9" x2="9" y2="15"></line><line x1="9" y1="9" x2="15" y2="15"></line></svg>`;
                        badge = `<span style="background: var(--error-bg); color: var(--error); padding: 0.25rem 0.5rem; border-radius: 4px; font-size: 0.75rem; font-weight: 600;">No Match</span>`;
                    }
                    
                    html += `
                        <div class="card requirement-card" style="display: flex; gap: 1rem; align-items: flex-start;">
                            <div style="flex-shrink: 0; margin-top: 0.25rem;">${icon}</div>
                            <div style="flex-grow: 1;">
                                <div style="display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 1rem;">
                                    <h4 style="margin: 0; font-size: 1.1rem;">${req.requirement}</h4>
                                    ${badge}
                                </div>
                                <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 1rem;">
                                    <div style="background: var(--bg-surface-elevated); padding: 1rem; border-radius: var(--radius-sm); border: 1px solid var(--border-color);">
                                        <div style="font-size: 0.75rem; text-transform: uppercase; letter-spacing: 0.05em; color: var(--text-muted); margin-bottom: 0.5rem;">Job Description</div>
                                        <div style="font-size: 0.875rem;">${req.jdEvidence ? req.jdEvidence.text : 'N/A'}</div>
                                    </div>
                                    <div style="background: var(--bg-surface-elevated); padding: 1rem; border-radius: var(--radius-sm); border: 1px solid var(--border-color);">
                                        <div style="font-size: 0.75rem; text-transform: uppercase; letter-spacing: 0.05em; color: var(--text-muted); margin-bottom: 0.5rem;">Resume Evidence</div>
                                        <div style="font-size: 0.875rem;">
                                            ${req.resumeEvidence && req.resumeEvidence.length > 0 ? req.resumeEvidence.map(e => `<div style="margin-bottom: 0.5rem;">"<i>${e.text}</i>"</div>`).join('') : '<span class="text-muted">No evidence found.</span>'}
                                        </div>
                                    </div>
                                </div>
                                ${req.note ? `<div class="mt-4 text-sm" style="background: var(--bg-surface); padding: 1rem; border-radius: var(--radius-sm); border: 1px dashed var(--border-color); display: flex; gap: 0.75rem; align-items: flex-start; color: var(--text-muted);">
                                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="flex-shrink: 0; margin-top: 0.1rem; color: var(--color-primary);"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="16" x2="12" y2="12"></line><line x1="12" y1="8" x2="12.01" y2="8"></line></svg>
                                    <div><strong style="color: var(--color-text);">Note:</strong> ${req.note}</div>
                                </div>` : ''}
                            </div>
                        </div>
                    `;
                });
                html += `</div>`;
            } else {
                html += `<div class="text-muted">No specific requirements analyzed for this section.</div>`;
            }
            
            html += `</div>`;
        });
        html += `</div>`;
    }
    
    container.innerHTML = html;
}

window.switchTab = function(secId) {
    document.querySelectorAll('.report-tab').forEach(el => {
        el.style.borderBottom = '2px solid transparent';
        el.style.fontWeight = 'normal';
        el.style.color = 'var(--text-muted)';
    });
    const activeTab = document.querySelector(`.report-tab[data-target="tab-${secId}"]`);
    if (activeTab) {
        activeTab.style.borderBottom = '2px solid var(--accent)';
        activeTab.style.fontWeight = '600';
        activeTab.style.color = 'var(--text-primary)';
    }
    
    document.querySelectorAll('.report-tab-content').forEach(el => {
        el.style.display = 'none';
    });
    const activeContent = document.getElementById(`tab-${secId}`);
    if (activeContent) {
        activeContent.style.display = 'block';
    }
};
