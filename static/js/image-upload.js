// Drag and drop functionality for image uploads
document.addEventListener('DOMContentLoaded', function() {
    const dropZone = document.querySelector('.upload-dropzone');
    const fileInput = document.getElementById('images');
    const uploadPreview = document.querySelector('.upload-preview');

    if (!dropZone || !fileInput) return;

    // Prevent default drag behaviors
    ['dragenter', 'dragover', 'dragleave', 'drop'].forEach(eventName => {
        dropZone.addEventListener(eventName, preventDefaults, false);
        document.body.addEventListener(eventName, preventDefaults, false);
    });

    // Highlight drop area when item is dragged over it
    ['dragenter', 'dragover'].forEach(eventName => {
        dropZone.addEventListener(eventName, highlight, false);
    });

    ['dragleave', 'drop'].forEach(eventName => {
        dropZone.addEventListener(eventName, unhighlight, false);
    });

    // Handle dropped files
    dropZone.addEventListener('drop', handleDrop, false);

    // Handle file selection via input
    fileInput.addEventListener('change', handleFiles, false);

    function preventDefaults(e) {
        e.preventDefault();
        e.stopPropagation();
    }

    function highlight() {
        dropZone.classList.add('drag-active');
    }

    function unhighlight() {
        dropZone.classList.remove('drag-active');
    }

    function handleDrop(e) {
        const dt = e.dataTransfer;
        const files = dt.files;
        handleFiles({ target: { files } });
    }

    function handleFiles(e) {
        const files = e.target.files;
        
        // Process each file
        for (let i = 0; i < files.length; i++) {
            const file = files[i];
            
            // Validate file type
            if (!file.type.match('image.*')) {
                alert('Please select only image files (JPG, PNG)');
                continue;
            }
            
            // Validate file size (max 5MB)
            if (file.size > 5 * 1024 * 1024) {
                alert(`File ${file.name} is too large (max 5MB)`);
                continue;
            }
            
            // Preview the image
            previewImage(file);
        }
    }

    function previewImage(file) {
        const reader = new FileReader();
        
        reader.onload = function(e) {
            const previewContainer = document.createElement('div');
            previewContainer.className = 'upload-preview-item';
            
            const img = document.createElement('img');
            img.src = e.target.result;
            img.alt = file.name;
            
            const fileName = document.createElement('small');
            fileName.textContent = truncateFileName(file.name, 20);
            
            previewContainer.appendChild(img);
            previewContainer.appendChild(fileName);
            
            if (uploadPreview) {
                uploadPreview.appendChild(previewContainer);
            }
        };
        
        reader.readAsDataURL(file);
    }

    function truncateFileName(name, maxLength) {
        if (name.length <= maxLength) return name;
        return name.substr(0, maxLength) + '...';
    }
});