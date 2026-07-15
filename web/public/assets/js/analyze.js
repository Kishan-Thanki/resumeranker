import { apiCall, logout, initAuth } from './api.js';
import { setupPage, showToast, setLoading } from './ui.js';

let resumeFile = null;
let jdFile = null;

let resumeUploadZone, resumeFileInput, resumeFileInfo, removeResumeBtn;
let jdUploadZone, jdFileInput, jdFileInfo, removeJdBtn;
let legalCheckbox, btnAnalyzeFit;
let newAnalysisView, analysisLoadingView;
let loadingStatusText, swapFilesContainer, btnSwapFiles;

const LOADING_STEPS = [
    "Extracting text from PDFs...",
    "Scrubbing Personal Identifiable Information (PII)...",
    "AI Engine analyzing core competencies...",
    "Evaluating against Job Description...",
    "Generating final report..."
];

function handleFileSelect(file, type) {
    if (!file) return;
    
    if (file.type !== 'application/pdf') {
        showToast('Please upload a PDF file.', true);
        return;
    }
    
    if (file.size > 5 * 1024 * 1024) {
        showToast('File too large. Max 5MB.', true);
        return;
    }
    
    if (type === 'resume') {
        resumeFile = file;
        resumeUploadZone.querySelector('svg').classList.add('hidden');
        resumeUploadZone.querySelector('h4').classList.add('hidden');
        resumeUploadZone.querySelector('p').classList.add('hidden');
        resumeFileInfo.classList.remove('hidden');
        resumeFileInfo.querySelector('.file-name').textContent = file.name;
    } else {
        jdFile = file;
        jdUploadZone.querySelector('svg').classList.add('hidden');
        jdUploadZone.querySelector('h4').classList.add('hidden');
        jdUploadZone.querySelector('p').classList.add('hidden');
        jdFileInfo.classList.remove('hidden');
        jdFileInfo.querySelector('.file-name').textContent = file.name;
    }
    
    checkSubmitState();
}

function removeFile(type, e) {
    if (e) e.stopPropagation();
    
    if (type === 'resume') {
        resumeFile = null;
        if (resumeFileInput) resumeFileInput.value = '';
        if (resumeFileInfo) resumeFileInfo.classList.add('hidden');
        if (resumeUploadZone) {
            resumeUploadZone.querySelector('svg').classList.remove('hidden');
            resumeUploadZone.querySelector('h4').classList.remove('hidden');
            resumeUploadZone.querySelector('p').classList.remove('hidden');
        }
    } else {
        jdFile = null;
        if (jdFileInput) jdFileInput.value = '';
        if (jdFileInfo) jdFileInfo.classList.add('hidden');
        if (jdUploadZone) {
            jdUploadZone.querySelector('svg').classList.remove('hidden');
            jdUploadZone.querySelector('h4').classList.remove('hidden');
            jdUploadZone.querySelector('p').classList.remove('hidden');
        }
    }
    checkSubmitState();
}

function checkSubmitState() {
    if (resumeFile && jdFile) {
        if (swapFilesContainer) swapFilesContainer.classList.remove('hidden');
    } else {
        if (swapFilesContainer) swapFilesContainer.classList.add('hidden');
    }

    if (resumeFile && jdFile && legalCheckbox && legalCheckbox.checked) {
        if (btnAnalyzeFit) btnAnalyzeFit.disabled = false;
    } else {
        if (btnAnalyzeFit) btnAnalyzeFit.disabled = true;
    }
}

function swapFiles() {
    if (!resumeFile || !jdFile) return;
    const tempFile = resumeFile;
    const tempName = resumeFile.name;
    
    // update resume to jd
    resumeFile = jdFile;
    resumeFileInfo.querySelector('.file-name').textContent = jdFile.name;
    
    // update jd to temp
    jdFile = tempFile;
    jdFileInfo.querySelector('.file-name').textContent = tempName;
}

function setupDragDrop(zone, type) {
    if (!zone) return;
    
    zone.addEventListener('dragover', (e) => {
        e.preventDefault();
        zone.classList.add('dragover');
    });
    
    zone.addEventListener('dragleave', () => {
        zone.classList.remove('dragover');
    });
    
    zone.addEventListener('drop', (e) => {
        e.preventDefault();
        zone.classList.remove('dragover');
        
        if (e.dataTransfer.files.length > 0) {
            handleFileSelect(e.dataTransfer.files[0], type);
        }
    });
    
    zone.addEventListener('click', () => {
        if (type === 'resume' && !resumeFile) {
            resumeFileInput.click();
        } else if (type === 'jd' && !jdFile) {
            jdFileInput.click();
        }
    });
}

function showView(viewName) {
    if (newAnalysisView) newAnalysisView.classList.add('hidden');
    if (analysisLoadingView) analysisLoadingView.classList.add('hidden');
    
    if (viewName === 'new') {
        if (newAnalysisView) newAnalysisView.classList.remove('hidden');
    } else if (viewName === 'loading') {
        if (analysisLoadingView) analysisLoadingView.classList.remove('hidden');
    }
}

async function submitAnalysis() {
    if (!resumeFile || !jdFile || !legalCheckbox.checked) return;
    
    showView('loading');
    setLoading(btnAnalyzeFit, true);
    
    let step = 0;
    if (loadingStatusText) loadingStatusText.textContent = LOADING_STEPS[step];
    
    const interval = setInterval(() => {
        step = (step + 1) % LOADING_STEPS.length;
        if (loadingStatusText) {
            loadingStatusText.textContent = LOADING_STEPS[step];
        }
    }, 4000);
    
    const formData = new FormData();
    formData.append('resume', resumeFile);
    formData.append('job_description', jdFile);
    
    try {
        const res = await apiCall('/dashboard/analyze/resume', {
            method: 'POST',
            body: formData
        });
        
        if (!res.ok && res.status !== 202) {
            const data = await res.json().catch(() => ({}));
            throw new Error(data.error || 'Analysis failed. Make sure PDFs are readable and contain text.');
        }
        
        const data = await res.json();
        const reqId = data.request_id || data.id;
        
        // Poll for completion
        const pollInterval = setInterval(async () => {
            try {
                const statusRes = await apiCall(`/dashboard/analyze/status?id=${reqId}`);
                if (!statusRes.ok) throw new Error("Status check failed");
                
                const statusData = await statusRes.json();
                
                if (statusData.status === "completed") {
                    clearInterval(pollInterval);
                    clearInterval(interval);
                    window.location.href = `/dashboard/report?id=${reqId}&res=${encodeURIComponent(resumeFile.name)}&jd=${encodeURIComponent(jdFile.name)}`;
                } else if (statusData.status === "failed") {
                    throw new Error(statusData.error || "Analysis failed during processing.");
                }
            } catch (e) {
                clearInterval(pollInterval);
                clearInterval(interval);
                showToast(e.message, true);
                showView('new');
                setLoading(btnAnalyzeFit, false);
            }
        }, 3000);
        
    } catch (err) {
        clearInterval(interval);
        showToast(err.message, true);
        showView('new');
        setLoading(btnAnalyzeFit, false);
    }
}

document.addEventListener('DOMContentLoaded', async () => {
    setupPage();
    await initAuth();
    
    resumeUploadZone = document.getElementById('resumeUploadZone');
    resumeFileInput = document.getElementById('resumeFileInput');
    resumeFileInfo = document.getElementById('resumeFileInfo');
    removeResumeBtn = document.getElementById('removeResumeBtn');
    
    jdUploadZone = document.getElementById('jdUploadZone');
    jdFileInput = document.getElementById('jdFileInput');
    jdFileInfo = document.getElementById('jdFileInfo');
    removeJdBtn = document.getElementById('removeJdBtn');
    
    legalCheckbox = document.getElementById('legalCheckbox');
    btnAnalyzeFit = document.getElementById('btnAnalyzeFit');
    
    newAnalysisView = document.getElementById('newAnalysisView');
    analysisLoadingView = document.getElementById('analysisLoadingView');
    
    loadingStatusText = document.getElementById('loadingStatusText');
    swapFilesContainer = document.getElementById('swapFilesContainer');
    btnSwapFiles = document.getElementById('btnSwapFiles');
    
    setupDragDrop(resumeUploadZone, 'resume');
    setupDragDrop(jdUploadZone, 'jd');
    
    if (resumeFileInput) {
        resumeFileInput.addEventListener('change', (e) => {
            if (e.target.files.length > 0) handleFileSelect(e.target.files[0], 'resume');
        });
    }
    
    if (jdFileInput) {
        jdFileInput.addEventListener('change', (e) => {
            if (e.target.files.length > 0) handleFileSelect(e.target.files[0], 'jd');
        });
    }
    
    if (removeResumeBtn) removeResumeBtn.addEventListener('click', (e) => removeFile('resume', e));
    if (removeJdBtn) removeJdBtn.addEventListener('click', (e) => removeFile('jd', e));
    
    if (legalCheckbox) legalCheckbox.addEventListener('change', checkSubmitState);
    if (btnAnalyzeFit) btnAnalyzeFit.addEventListener('click', submitAnalysis);
    
    if (btnSwapFiles) btnSwapFiles.addEventListener('click', swapFiles);
    
    showView('new');
});
