import './style.css'

// Helper function to handle the form submission, API call, and file download
async function handleConversion(
  formEvent: SubmitEvent,
  inputElement: HTMLInputElement,
  statusElement: HTMLParagraphElement,
  apiUrl: string,
  downloadFilename: string
) {
  formEvent.preventDefault();
  statusElement.classList.remove('hidden', 'text-red-600', 'text-green-600');
  statusElement.classList.add('text-gray-500');

  if (!inputElement.files || inputElement.files.length === 0) {
    statusElement.textContent = 'Please select a file first.';
    statusElement.classList.add('text-red-600');
    return;
  }

  const file = inputElement.files[0];
  const formData = new FormData();
  formData.append('file', file);

  try {
    statusElement.textContent = 'Uploading and converting...';
    
    // In dev mode, this hits the Vite proxy, in prod it hits the Go server directly.
    const response = await fetch(apiUrl, {
      method: 'POST',
      body: formData,
    });

    if (!response.ok) {
      const errText = await response.text();
      throw new Error(errText || `Server error: ${response.status}`);
    }

    // Get the converted file blob
    const blob = await response.blob();
    
    // Create a temporary link to trigger the download
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.style.display = 'none';
    a.href = url;
    a.download = downloadFilename;
    
    document.body.appendChild(a);
    a.click();
    
    // Cleanup
    window.URL.revokeObjectURL(url);
    document.body.removeChild(a);

    statusElement.textContent = 'Conversion successful! File downloaded.';
    statusElement.classList.replace('text-gray-500', 'text-green-600');
  } catch (error) {
    console.error('Conversion error:', error);
    statusElement.textContent = `Error: ${error instanceof Error ? error.message : 'Unknown error occurred'}`;
    statusElement.classList.replace('text-gray-500', 'text-red-600');
  }
}

// Setup Event Listeners
document.addEventListener('DOMContentLoaded', () => {
  const imageForm = document.getElementById('image-form') as HTMLFormElement;
  const imageInput = document.getElementById('image-input') as HTMLInputElement;
  const imageStatus = document.getElementById('image-status') as HTMLParagraphElement;

  const documentForm = document.getElementById('document-form') as HTMLFormElement;
  const documentInput = document.getElementById('document-input') as HTMLInputElement;
  const documentStatus = document.getElementById('document-status') as HTMLParagraphElement;

  if (imageForm) {
    imageForm.addEventListener('submit', (e) => {
      handleConversion(e, imageInput, imageStatus, '/convert/image', 'converted.png');
    });
  }

  if (documentForm) {
    documentForm.addEventListener('submit', (e) => {
      handleConversion(e, documentInput, documentStatus, '/convert/document', 'converted.pdf');
    });
  }
});
