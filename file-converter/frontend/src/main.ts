import './style.css'

const fromSelect = document.getElementById('from-type') as HTMLSelectElement;
const toSelect = document.getElementById('to-type') as HTMLSelectElement;
const fileInput = document.getElementById('file-input') as HTMLInputElement;
const convertForm = document.getElementById('convert-form') as HTMLFormElement;
const statusMessage = document.getElementById('status-message') as HTMLParagraphElement;

function updateFileInputAccept() {
  const selectedFrom = fromSelect.value;
  fileInput.accept = `.${selectedFrom}`;
}

// Initial setup
updateFileInputAccept();

// Event listeners
fromSelect.addEventListener('change', updateFileInputAccept);

convertForm.addEventListener('submit', async (e) => {
  e.preventDefault();
  
  statusMessage.classList.remove('hidden', 'text-red-600', 'text-green-600');
  statusMessage.classList.add('text-gray-500');

  if (!fileInput.files || fileInput.files.length === 0) {
    statusMessage.textContent = 'Please select a file first.';
    statusMessage.classList.add('text-red-600');
    return;
  }

  const file = fileInput.files[0];
  const fromType = fromSelect.value;
  const toType = toSelect.value;

  if (fromType === toType) {
    statusMessage.textContent = 'Source and target formats cannot be the same.';
    statusMessage.classList.add('text-red-600');
    return;
  }

  const formData = new FormData();
  formData.append('file', file);
  formData.append('fromType', fromType);
  formData.append('toType', toType);

  try {
    statusMessage.textContent = `Converting ${fromType.toUpperCase()} to ${toType.toUpperCase()}...`;
    
    const response = await fetch('/convert', {
      method: 'POST',
      body: formData,
    });

    if (!response.ok) {
      const errText = await response.text();
      throw new Error(errText || `Server error: ${response.status}`);
    }

    const blob = await response.blob();
    
    // Create download link
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.style.display = 'none';
    a.href = url;
    a.download = `converted.${toType}`;
    
    document.body.appendChild(a);
    a.click();
    
    window.URL.revokeObjectURL(url);
    document.body.removeChild(a);

    statusMessage.textContent = `Success! File converted to ${toType.toUpperCase()}.`;
    statusMessage.classList.replace('text-gray-500', 'text-green-600');
  } catch (error) {
    console.error('Conversion error:', error);
    statusMessage.textContent = `Error: ${error instanceof Error ? error.message : 'Unknown error occurred'}`;
    statusMessage.classList.replace('text-gray-500', 'text-red-600');
  }
});
