import './style.css'

// ----------------------------------------------------
// UI STATE & NAVIGATION
// ----------------------------------------------------
const btnFileConverter = document.getElementById('btn-file-converter') as HTMLButtonElement;
const btnMeasureConverter = document.getElementById('btn-measure-converter') as HTMLButtonElement;
const btnTimeConverter = document.getElementById('btn-time-converter') as HTMLButtonElement;

const fileConverterView = document.getElementById('file-converter-view') as HTMLDivElement;
const measureConverterView = document.getElementById('measure-converter-view') as HTMLDivElement;
const timeConverterView = document.getElementById('time-converter-view') as HTMLDivElement;

function setActiveView(view: 'file' | 'measure' | 'time') {
  // Hide all
  fileConverterView.classList.add('hidden');
  measureConverterView.classList.add('hidden');
  timeConverterView.classList.add('hidden');
  
  // Reset buttons
  btnFileConverter.classList.remove('bg-indigo-800', 'text-white');
  btnFileConverter.classList.add('text-indigo-200');
  
  btnMeasureConverter.classList.remove('bg-indigo-800', 'text-white');
  btnMeasureConverter.classList.add('text-indigo-200');
  
  btnTimeConverter.classList.remove('bg-indigo-800', 'text-white');
  btnTimeConverter.classList.add('text-indigo-200');

  // Activate selected
  if (view === 'file') {
    fileConverterView.classList.remove('hidden');
    btnFileConverter.classList.replace('text-indigo-200', 'text-white');
    btnFileConverter.classList.add('bg-indigo-800');
  } else if (view === 'measure') {
    measureConverterView.classList.remove('hidden');
    btnMeasureConverter.classList.replace('text-indigo-200', 'text-white');
    btnMeasureConverter.classList.add('bg-indigo-800');
  } else if (view === 'time') {
    timeConverterView.classList.remove('hidden');
    btnTimeConverter.classList.replace('text-indigo-200', 'text-white');
    btnTimeConverter.classList.add('bg-indigo-800');
  }
}

btnFileConverter.addEventListener('click', () => setActiveView('file'));
btnMeasureConverter.addEventListener('click', () => setActiveView('measure'));
btnTimeConverter.addEventListener('click', () => setActiveView('time'));

// ----------------------------------------------------
// FILE CONVERTER LOGIC
// ----------------------------------------------------
const toSelect = document.getElementById('to-type') as HTMLSelectElement;
const fileInput = document.getElementById('file-input') as HTMLInputElement;
const convertForm = document.getElementById('convert-form') as HTMLFormElement;
const statusMessage = document.getElementById('status-message') as HTMLParagraphElement;
const detectionBox = document.getElementById('detection-box') as HTMLDivElement;
const detectedFormatText = document.getElementById('detected-format-text') as HTMLSpanElement;

let detectedFromType = '';

fileInput.addEventListener('change', () => {
  if (fileInput.files && fileInput.files.length > 0) {
    const file = fileInput.files[0];
    const ext = file.name.split('.').pop()?.toLowerCase();
    
    if (ext) {
      detectedFromType = ext;
      detectedFormatText.textContent = ext.toUpperCase() + ' (Detected)';
      detectionBox.classList.replace('bg-gray-100', 'bg-indigo-50');
      detectionBox.classList.replace('border-gray-200', 'border-indigo-200');
      detectedFormatText.classList.replace('text-gray-500', 'text-indigo-700');
    } else {
      resetDetectionBox();
    }
  } else {
    resetDetectionBox();
  }
});

function resetDetectionBox() {
  detectedFromType = '';
  detectedFormatText.textContent = 'Auto-detected on upload';
  detectionBox.classList.replace('bg-indigo-50', 'bg-gray-100');
  detectionBox.classList.replace('border-indigo-200', 'border-gray-200');
  detectedFormatText.classList.replace('text-indigo-700', 'text-gray-500');
}

convertForm.addEventListener('submit', async (e) => {
  e.preventDefault();
  
  statusMessage.classList.remove('hidden', 'text-red-600', 'text-green-600');
  statusMessage.classList.add('text-gray-500');

  if (!fileInput.files || fileInput.files.length === 0 || !detectedFromType) {
    statusMessage.textContent = 'Please select a valid file first.';
    statusMessage.classList.add('text-red-600');
    return;
  }

  const file = fileInput.files[0];
  const toType = toSelect.value;

  if (detectedFromType === toType) {
    statusMessage.textContent = 'Source and target formats cannot be the same.';
    statusMessage.classList.add('text-red-600');
    return;
  }

  const formData = new FormData();
  formData.append('file', file);
  formData.append('fromType', detectedFromType);
  formData.append('toType', toType);

  try {
    statusMessage.textContent = `Converting ${detectedFromType.toUpperCase()} to ${toType.toUpperCase()}...`;
    
    const isLocal = window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1';
    const apiUrl = isLocal 
      ? 'http://localhost:8080/convert' 
      : 'https://utils.api.srilakshmiretail.in/convert';
      
    const response = await fetch(apiUrl, {
      method: 'POST',
      body: formData,
    });

    if (!response.ok) {
      const errText = await response.text();
      throw new Error(errText || `Server error: ${response.status}`);
    }

    const blob = await response.blob();
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

// ----------------------------------------------------
// MEASUREMENT CONVERTER LOGIC
// ----------------------------------------------------
const unitsData: Record<string, string[]> = {
  length: ["meters", "kilometers", "centimeters", "millimeters", "miles", "yards", "feet", "inches"],
  weight: ["kilograms", "grams", "milligrams", "pounds", "ounces"],
  volume: ["liters", "milliliters", "gallons", "quarts", "pints", "fluid_ounces"],
  area: ["square_meters", "square_kilometers", "hectares", "acres", "square_feet", "square_miles"],
  time: ["seconds", "minutes", "hours", "days", "weeks"],
  temperature: ["celsius", "fahrenheit", "kelvin"],
  speed: ["meters_per_second", "kilometers_per_hour", "miles_per_hour", "feet_per_second", "knots"]
};

const measureCategory = document.getElementById('measure-category') as HTMLSelectElement;
const measureFrom = document.getElementById('measure-from') as HTMLSelectElement;
const measureTo = document.getElementById('measure-to') as HTMLSelectElement;
const measureValue = document.getElementById('measure-value') as HTMLInputElement;
const measureForm = document.getElementById('measure-form') as HTMLFormElement;
const measureResultBox = document.getElementById('measure-result-box') as HTMLDivElement;
const measureResultText = document.getElementById('measure-result-text') as HTMLSpanElement;
const measureStatusMessage = document.getElementById('measure-status-message') as HTMLParagraphElement;

function populateUnits(category: string) {
  measureFrom.innerHTML = '';
  measureTo.innerHTML = '';
  
  const units = unitsData[category] || [];
  units.forEach(unit => {
    // format string (e.g., square_meters -> Square Meters)
    const label = unit.split('_').map(w => w.charAt(0).toUpperCase() + w.slice(1)).join(' ');
    
    measureFrom.add(new Option(label, unit));
    measureTo.add(new Option(label, unit));
  });

  // Default select the second option for "To" if available
  if (units.length > 1) {
    measureTo.selectedIndex = 1;
  }
}

// Initial populate
populateUnits(measureCategory.value);

measureCategory.addEventListener('change', () => {
  populateUnits(measureCategory.value);
  measureResultBox.classList.add('hidden');
});

measureForm.addEventListener('submit', async (e) => {
  e.preventDefault();
  
  measureStatusMessage.classList.remove('hidden', 'text-red-600', 'text-green-600');
  measureStatusMessage.classList.add('text-gray-500');
  measureStatusMessage.textContent = 'Converting...';
  measureResultBox.classList.add('hidden');

  const payload = {
    category: measureCategory.value,
    fromUnit: measureFrom.value,
    toUnit: measureTo.value,
    value: parseFloat(measureValue.value)
  };

  try {
    const isLocal = window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1';
    const apiUrl = isLocal 
      ? 'http://localhost:8080/convert-measurement' 
      : 'https://utils.api.srilakshmiretail.in/convert-measurement';
      
    const response = await fetch(apiUrl, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(payload),
    });

    if (!response.ok) {
      const errText = await response.text();
      throw new Error(errText || `Server error: ${response.status}`);
    }

    const data = await response.json();
    
    // Format the number to remove unnecessary trailing decimals (e.g. 50.853015 -> maybe fix if needed)
    const formattedResult = Number.isInteger(data.result) ? data.result : Number(data.result.toFixed(6));
    
    measureResultText.textContent = `${formattedResult}`;
    measureResultBox.classList.remove('hidden');
    measureStatusMessage.classList.add('hidden');
    
  } catch (error) {
    console.error('Measurement conversion error:', error);
    measureStatusMessage.textContent = `Error: ${error instanceof Error ? error.message : 'Unknown error occurred'}`;
    measureStatusMessage.classList.replace('text-gray-500', 'text-red-600');
  }
});

// ----------------------------------------------------
// TIME CONVERTER LOGIC
// ----------------------------------------------------
const timeForm = document.getElementById('time-form') as HTMLFormElement;
const timeDateInput = document.getElementById('time-date') as HTMLInputElement;
const timeHourInput = document.getElementById('time-hour') as HTMLSelectElement;
const timeMinuteInput = document.getElementById('time-minute') as HTMLSelectElement;
const timeAmpmInput = document.getElementById('time-ampm') as HTMLSelectElement;
const timeSourceTzInput = document.getElementById('time-source-tz') as HTMLSelectElement;
const timeDestTzInput = document.getElementById('time-dest-tz') as HTMLSelectElement;
const timeResultBox = document.getElementById('time-result-box') as HTMLDivElement;
const timeResultLocal = document.getElementById('time-result-local') as HTMLSpanElement;
const timeResultZone = document.getElementById('time-result-zone') as HTMLSpanElement;
const timeWarning = document.getElementById('time-warning') as HTMLParagraphElement;
const timeStatusMessage = document.getElementById('time-status-message') as HTMLParagraphElement;

const timezones = [
  "UTC", "America/New_York", "America/Chicago", "America/Denver", "America/Los_Angeles", 
  "America/Phoenix", "America/Anchorage", "America/Honolulu", "America/Sao_Paulo", 
  "America/Argentina/Buenos_Aires", "America/Bogota", "Europe/London", "Europe/Paris", 
  "Europe/Berlin", "Europe/Rome", "Europe/Moscow", "Africa/Cairo", "Africa/Johannesburg", 
  "Africa/Lagos", "Asia/Dubai", "Asia/Kolkata", "Asia/Dhaka", "Asia/Bangkok", "Asia/Singapore", 
  "Asia/Hong_Kong", "Asia/Shanghai", "Asia/Tokyo", "Asia/Seoul", "Australia/Sydney", 
  "Australia/Melbourne", "Australia/Brisbane", "Australia/Perth", "Pacific/Auckland", "Pacific/Fiji"
];

timezones.forEach(tz => {
  timeSourceTzInput.add(new Option(tz, tz));
  timeDestTzInput.add(new Option(tz, tz));
});

for (let h = 1; h <= 12; h++) {
  const hr = h.toString().padStart(2, '0');
  timeHourInput.add(new Option(hr, hr));
}
for (let m = 0; m <= 59; m++) {
  const min = m.toString().padStart(2, '0');
  timeMinuteInput.add(new Option(min, min));
}

timeSourceTzInput.value = "America/New_York";
timeDestTzInput.value = "Asia/Tokyo";

timeForm.addEventListener('submit', async (e) => {
  e.preventDefault();
  
  timeStatusMessage.classList.remove('hidden', 'text-red-600', 'text-green-600');
  timeStatusMessage.classList.add('text-gray-500');
  timeStatusMessage.textContent = 'Converting...';
  timeResultBox.classList.add('hidden');
  timeWarning.classList.add('hidden');

  const dateValue = timeDateInput.value; // e.g. "2024-11-03"
  if (!dateValue) return;

  const dateParts = dateValue.split('-');
  
  let hour = parseInt(timeHourInput.value, 10);
  const minute = parseInt(timeMinuteInput.value, 10);
  const ampm = timeAmpmInput.value;
  
  // Convert to 24hr for the payload request internally
  if (ampm === "PM" && hour !== 12) hour += 12;
  if (ampm === "AM" && hour === 12) hour = 0;
  
  const payload = {
    year: parseInt(dateParts[0], 10),
    month: parseInt(dateParts[1], 10),
    day: parseInt(dateParts[2], 10),
    hour: hour,
    minute: minute,
    second: 0,
    source_tz: timeSourceTzInput.value,
    dest_tz: timeDestTzInput.value,
    ambiguous_policy: "first",
    non_existent_policy: "forward"
  };

  try {
    const isLocal = window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1';
    const apiUrl = isLocal 
      ? 'http://localhost:8080/convert-time' 
      : 'https://utils.api.srilakshmiretail.in/convert-time';
      
    const response = await fetch(apiUrl, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(payload),
    });

    if (!response.ok) {
      const errText = await response.text();
      throw new Error(errText || `Server error: ${response.status}`);
    }

    const data = await response.json();
    
    // Parse local time for 12-hour AM/PM format
    const destLocalParts = data.dest_time_local.split('T');
    const destDate = destLocalParts[0];
    const timeWithOffset = destLocalParts[1];
    
    const [hourStr, minStr] = timeWithOffset.split(':');
    let hour = parseInt(hourStr, 10);
    const ampm = hour >= 12 ? 'PM' : 'AM';
    hour = hour % 12;
    if (hour === 0) hour = 12;
    
    const destTime12 = `${hour.toString().padStart(2, '0')}:${minStr} ${ampm}`;
    
    timeResultLocal.textContent = `${destDate} ${destTime12}`;
    
    let zoneText = `${data.dest_zone_name} (UTC ${data.dest_offset})`;
    if (data.is_next_day) zoneText += ' • Next Day';
    if (data.is_prev_day) zoneText += ' • Previous Day';
    
    timeResultZone.textContent = zoneText;
    
    if (data.warning) {
      timeWarning.textContent = data.warning;
      timeWarning.classList.remove('hidden');
    }

    timeResultBox.classList.remove('hidden');
    timeStatusMessage.classList.add('hidden');
    
  } catch (error) {
    console.error('Time conversion error:', error);
    timeStatusMessage.textContent = `Error: ${error instanceof Error ? error.message : 'Unknown error occurred'}`;
    timeStatusMessage.classList.replace('text-gray-500', 'text-red-600');
  }
});
