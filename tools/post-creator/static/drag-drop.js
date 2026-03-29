(function() {
    var dropZone = document.getElementById('drop-zone');
    var fileInput = document.getElementById('image-input');
    var modal = document.getElementById('image-modal');

    if (!dropZone) return;

    dropZone.addEventListener('click', function() { fileInput.click(); });
    fileInput.addEventListener('change', function(e) { handleFiles(e.target.files); });

    dropZone.addEventListener('dragover', function(e) {
        e.preventDefault();
        dropZone.classList.add('drag-over');
    });
    dropZone.addEventListener('dragleave', function() {
        dropZone.classList.remove('drag-over');
    });
    dropZone.addEventListener('drop', function(e) {
        e.preventDefault();
        dropZone.classList.remove('drag-over');
        handleFiles(e.dataTransfer.files);
    });

    function handleFiles(files) {
        if (files.length === 0) return;
        var file = files[0];
        if (!file.type.startsWith('image/')) return;

        var formData = new FormData();
        formData.append('image', file);

        fetch('/api/images/upload', { method: 'POST', body: formData })
            .then(function(r) { return r.json(); })
            .then(function(data) { showModal(data); })
            .catch(function(err) { console.error('Upload failed:', err); });
    }

    function showModal(data) {
        document.getElementById('modal-preview').src = '/images/' + data.filename;
        document.getElementById('shortcode-type').value = data.recommended_shortcode;
        document.getElementById('caption-input').value = '';
        modal.dataset.filename = data.filename;
        modal.style.display = 'flex';
    }

    document.getElementById('copy-shortcode').addEventListener('click', function() {
        var type = document.getElementById('shortcode-type').value;
        var caption = document.getElementById('caption-input').value;
        var filename = modal.dataset.filename;
        var src = '/images/' + filename;

        var shortcode;
        if (caption) {
            shortcode = '{{< ' + type + ' src="' + src + '" caption="' + caption + '" >}}';
        } else {
            shortcode = '{{< ' + type + ' src="' + src + '" >}}';
        }

        var btn = document.getElementById('copy-shortcode');
        navigator.clipboard.writeText(shortcode).then(function() {
            btn.textContent = 'Copied!';
            setTimeout(function() { btn.textContent = 'Copy to Clipboard'; }, 1500);
        });
    });

    document.getElementById('close-modal').addEventListener('click', function() {
        modal.style.display = 'none';
    });
})();
