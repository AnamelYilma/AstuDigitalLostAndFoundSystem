// Drag and drop functionality for image uploads
document.addEventListener('DOMContentLoaded', function() {
    const dropZone = document.querySelector('.upload-dropzone');
    const fileInput = document.getElementById('images');
    const uploadPreview = document.querySelector('.upload-preview');
    const MAX_FILES = 5;

    if (!dropZone || !fileInput) return;

    let selection = new DataTransfer();

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
        handleFiles({ target: { files: dt.files } });
    }

    function handleFiles(e) {
        const files = e.target.files;
        if (!files || files.length === 0) return;

        selection = new DataTransfer();

        let added = 0;
        for (let i = 0; i < files.length; i++) {
            const file = files[i];

            if (!file.type.match('image.*')) {
                alert('Please select only image files (JPG, PNG)');
                continue;
            }

            if (file.size > 5 * 1024 * 1024) {
                alert(`File ${file.name} is too large (max 5MB)`);
                continue;
            }

            if (selection.files.length >= MAX_FILES) {
                alert(`You can upload up to ${MAX_FILES} images.`);
                break;
            }

            selection.items.add(file);
            added++;
        }

        if (added > 0) {
            syncSelection();
        }
    }

    function syncSelection() {
        fileInput.files = selection.files;
        renderPreview();
    }

    function renderPreview() {
        if (!uploadPreview) return;
        uploadPreview.innerHTML = '';

        if (selection.files.length === 0) return;

        Array.from(selection.files).forEach(file => {
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

                uploadPreview.appendChild(previewContainer);
            };
            reader.readAsDataURL(file);
        });
    }

    function truncateFileName(name, maxLength) {
        if (name.length <= maxLength) return name;
        return name.substr(0, maxLength) + '...';
    }
});
