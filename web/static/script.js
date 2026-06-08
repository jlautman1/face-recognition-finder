// Source tab switching
let activeSource = 'upload';

document.querySelectorAll('.source-tab').forEach(tab => {
    tab.addEventListener('click', function() {
        document.querySelectorAll('.source-tab').forEach(t => t.classList.remove('active'));
        document.querySelectorAll('.source-panel').forEach(p => p.classList.remove('active'));
        this.classList.add('active');
        activeSource = this.dataset.source;
        document.getElementById('panel-' + activeSource).classList.add('active');
    });
});

// File upload handlers
document.getElementById('zipFile').addEventListener('change', function(e) {
    const fileName = e.target.files[0]?.name || 'No file selected';
    document.getElementById('zipFileName').textContent = fileName;
});

document.getElementById('refImages').addEventListener('change', function(e) {
    const files = e.target.files;
    const fileCount = files.length;
    document.getElementById('refFileName').textContent = 
        fileCount > 0 ? `${fileCount} file${fileCount > 1 ? 's' : ''} selected` : 'No files selected';
    
    // Show image previews
    const previewContainer = document.getElementById('imagePreview');
    previewContainer.innerHTML = '';
    
    Array.from(files).slice(0, 10).forEach(file => {
        const reader = new FileReader();
        reader.onload = function(e) {
            const img = document.createElement('img');
            img.src = e.target.result;
            img.className = 'preview-image';
            previewContainer.appendChild(img);
        };
        reader.readAsDataURL(file);
    });
});

// Threshold slider
document.getElementById('threshold').addEventListener('input', function(e) {
    document.getElementById('thresholdValue').textContent = e.target.value;
});

// Form submission
document.getElementById('uploadForm').addEventListener('submit', async function(e) {
    e.preventDefault();
    
    const submitBtn = document.getElementById('submitBtn');

    // Validate based on active source
    if (activeSource === 'upload') {
        const zipFile = document.getElementById('zipFile').files[0];
        if (!zipFile) { alert('Please select a ZIP file.'); return; }
    } else if (activeSource === 'url') {
        const url = document.getElementById('sourceUrl').value.trim();
        if (!url) { alert('Please enter a download URL.'); return; }
    } else if (activeSource === 'local') {
        const path = document.getElementById('localPath').value.trim();
        if (!path) { alert('Please enter a local path.'); return; }
    }

    const refFiles = document.getElementById('refImages').files;
    if (refFiles.length === 0) { alert('Please upload at least one reference photo.'); return; }

    submitBtn.disabled = true;
    submitBtn.innerHTML = '<span>' + (activeSource === 'url' ? 'Submitting...' : 'Uploading...') + '</span>';
    
    const formData = new FormData();
    formData.append('sourceType', activeSource);
    formData.append('threshold', document.getElementById('threshold').value);

    // Add source-specific data
    if (activeSource === 'upload') {
        formData.append('zipFile', document.getElementById('zipFile').files[0]);
    } else if (activeSource === 'url') {
        formData.append('sourceUrl', document.getElementById('sourceUrl').value.trim());
    } else if (activeSource === 'local') {
        formData.append('localPath', document.getElementById('localPath').value.trim());
    }

    // Add reference images
    for (const file of refFiles) {
        formData.append('refImages', file);
    }
    
    try {
        const response = await fetch('/upload', {
            method: 'POST',
            body: formData
        });
        
        if (!response.ok) {
            throw new Error('Upload failed');
        }
        
        const result = await response.json();
        
        // Hide upload section, show processing
        document.getElementById('uploadSection').classList.add('hidden');
        document.getElementById('processingSection').classList.remove('hidden');
        
        // Start polling for status
        pollStatus(result.job_id);
        
    } catch (error) {
        alert('Upload failed: ' + error.message);
        submitBtn.disabled = false;
        submitBtn.innerHTML = `
            <svg width="20" height="20" viewBox="0 0 20 20" fill="none">
                <path d="M10 3L3 10L10 17M3 10H17" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
            Start Processing
        `;
    }
});

let statusInterval;

function pollStatus(jobId) {
    statusInterval = setInterval(async () => {
        try {
            const response = await fetch(`/status/${jobId}`);
            const status = await response.json();
            
            updateStatus(status);
            
            if (status.status === 'completed') {
                clearInterval(statusInterval);
                showResults(jobId, status);
            } else if (status.status === 'error') {
                clearInterval(statusInterval);
                showError(status.message);
            }
        } catch (error) {
            console.error('Status check failed:', error);
        }
    }, 2000); // Poll every 2 seconds
}

function updateStatus(status) {
    document.getElementById('statusTitle').textContent =
        status.status === 'processing' ? 'Processing Your Images...' : status.message;
    document.getElementById('statusMessage').textContent = status.message;
    document.getElementById('progressBar').style.width = status.progress + '%';

    if (status.total > 0) {
        document.getElementById('scannedCount').textContent = status.scanned + '/' + status.total;
    }

    document.getElementById('matchCount').textContent = status.matches || 0;
}

function showResults(jobId, status) {
    document.getElementById('processingSection').classList.add('hidden');
    document.getElementById('resultsSection').classList.remove('hidden');
    
    document.getElementById('totalMatches').textContent = status.matches;
    document.getElementById('downloadBtn').href = `/download/${jobId}`;
}

function showError(message) {
    document.getElementById('processingSection').classList.add('hidden');
    alert('Processing failed: ' + message);
    location.reload();
}
