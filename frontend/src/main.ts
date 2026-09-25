import './style.css'

// ----------------------------------------------------
// UI STATE & NAVIGATION
// ----------------------------------------------------
const btnFileConverter = document.getElementById('btn-file-converter') as HTMLButtonElement;
const btnMeasureConverter = document.getElementById('btn-measure-converter') as HTMLButtonElement;
const btnHrCalculator = document.getElementById('btn-hr-calculator') as HTMLButtonElement | null;
const btnTimeConverter = document.getElementById('btn-time-converter') as HTMLButtonElement;
const btnRailwayConverter = document.getElementById('btn-railway-converter') as HTMLButtonElement;

// PDF dropdown elements
const btnPdfMenu = document.getElementById('btn-pdf-menu') as HTMLButtonElement;
const pdfMenuDropdown = document.getElementById('pdf-menu-dropdown') as HTMLDivElement;
const pdfMenuIcon = document.getElementById('pdf-menu-icon') as SVGSVGElement;
const btnPdfSizeWork = document.getElementById('btn-pdf-size-work') as HTMLButtonElement;

const btnTimeMenu = document.getElementById('btn-time-menu') as HTMLButtonElement;
const timeMenuDropdown = document.getElementById('time-menu-dropdown') as HTMLDivElement;
const timeMenuIcon = document.getElementById('time-menu-icon') as SVGSVGElement;

const btnMeasureMenu = document.getElementById('btn-measure-menu') as HTMLButtonElement | null;
const measureMenuDropdown = document.getElementById('measure-menu-dropdown') as HTMLDivElement | null;
const measureMenuIcon = document.getElementById('measure-menu-icon') as SVGSVGElement | null;

const btnHrMenu = document.getElementById('btn-hr-menu') as HTMLButtonElement | null;
const hrMenuDropdown = document.getElementById('hr-menu-dropdown') as HTMLDivElement | null;
const hrMenuIcon = document.getElementById('hr-menu-icon') as SVGSVGElement | null;

const fileConverterView = document.getElementById('file-converter-view') as HTMLDivElement;
const measureConverterView = document.getElementById('measure-converter-view') as HTMLDivElement;
const hrCalculatorView = document.getElementById('hr-calculator-view') as HTMLDivElement;
const timeConverterView = document.getElementById('time-converter-view') as HTMLDivElement;
const railwayConverterView = document.getElementById('railway-converter-view') as HTMLDivElement;
const pdfConverterView = document.getElementById('pdf-converter-view') as HTMLDivElement;

let timeMenuOpen = false;
let measureMenuOpen = false;
let hrMenuOpen = false;
let pdfMenuOpen = false;

// ----------------------------------------------------
// MOBILE SIDEBAR TOGGLE
// ----------------------------------------------------
const sidebar = document.getElementById('sidebar') as HTMLElement | null;
const mobileMenuBtn = document.getElementById('mobile-menu-btn') as HTMLButtonElement | null;
const mobileMenuClose = document.getElementById('mobile-menu-close') as HTMLButtonElement | null;
const mobileOverlay = document.getElementById('mobile-overlay') as HTMLDivElement | null;

function openMobileMenu() {
  if (!sidebar) return;
  sidebar.classList.remove('-translate-x-full');
  sidebar.classList.add('translate-x-0');
  mobileOverlay?.classList.remove('hidden');
}

function closeMobileMenu() {
  if (!sidebar) return;
  sidebar.classList.add('-translate-x-full');
  sidebar.classList.remove('translate-x-0');
  mobileOverlay?.classList.add('hidden');
}

mobileMenuBtn?.addEventListener('click', openMobileMenu);
mobileMenuClose?.addEventListener('click', closeMobileMenu);
mobileOverlay?.addEventListener('click', closeMobileMenu);

// Auto-close the mobile menu when a navigation button is clicked
// so the panel is immediately visible on phones
document.querySelectorAll('#sidebar nav button').forEach(btn => {
  btn.addEventListener('click', () => {
    if (window.innerWidth < 768) {
      setTimeout(closeMobileMenu, 150);
    }
  });
});

// ----------------------------------------------------
// PDF MENU TOGGLE
// ----------------------------------------------------
btnPdfMenu.addEventListener('click', () => {
  pdfMenuOpen = !pdfMenuOpen;
  if (pdfMenuOpen) {
    pdfMenuDropdown.classList.remove('hidden');
    pdfMenuDropdown.classList.add('flex');
    pdfMenuIcon.classList.add('rotate-180');
  } else {
    pdfMenuDropdown.classList.add('hidden');
    pdfMenuDropdown.classList.remove('flex');
    pdfMenuIcon.classList.remove('rotate-180');
  }
});

// ----------------------------------------------------
// TIME MENU TOGGLE
// ----------------------------------------------------
btnTimeMenu.addEventListener('click', () => {
  timeMenuOpen = !timeMenuOpen;
  if (timeMenuOpen) {
    timeMenuDropdown.classList.remove('hidden');
    timeMenuDropdown.classList.add('flex');
    timeMenuIcon.classList.add('rotate-180');
  } else {
    timeMenuDropdown.classList.add('hidden');
    timeMenuDropdown.classList.remove('flex');
    timeMenuIcon.classList.remove('rotate-180');
  }
});

// ----------------------------------------------------
// MEASURE MENU TOGGLE
// ----------------------------------------------------
if (btnMeasureMenu && measureMenuDropdown && measureMenuIcon) {
  btnMeasureMenu.addEventListener('click', () => {
    measureMenuOpen = !measureMenuOpen;
    if (measureMenuOpen) {
      measureMenuDropdown.classList.remove('hidden');
      measureMenuDropdown.classList.add('flex');
      measureMenuIcon.classList.add('rotate-180');
    } else {
      measureMenuDropdown.classList.add('hidden');
      measureMenuDropdown.classList.remove('flex');
      measureMenuIcon.classList.remove('rotate-180');
    }
  });
}

// ----------------------------------------------------
// HR MENU TOGGLE
// ----------------------------------------------------
if (btnHrMenu && hrMenuDropdown && hrMenuIcon) {
  btnHrMenu.addEventListener('click', () => {
    hrMenuOpen = !hrMenuOpen;
    if (hrMenuOpen) {
      hrMenuDropdown.classList.remove('hidden');
      hrMenuDropdown.classList.add('flex');
      hrMenuIcon.classList.add('rotate-180');
    } else {
      hrMenuDropdown.classList.add('hidden');
      hrMenuDropdown.classList.remove('flex');
      hrMenuIcon.classList.remove('rotate-180');
    }
  });
}

// ----------------------------------------------------
// VIEW SWITCHER
// ----------------------------------------------------
function setActiveView(view: 'file' | 'pdf' | 'measure' | 'time' | 'railway' | 'hr' | 'age') {
  // Hide all views first
  fileConverterView.classList.add('hidden');
  pdfConverterView.classList.add('hidden');
  measureConverterView.classList.add('hidden');
  hrCalculatorView.classList.add('hidden');
  timeConverterView.classList.add('hidden');
  railwayConverterView.classList.add('hidden');
  ageCalculatorPanel.classList.add('hidden');

  // Reset all sidebar buttons
  btnFileConverter.classList.remove('bg-indigo-800', 'text-white');
  btnFileConverter.classList.add('text-indigo-200');

  btnPdfSizeWork.classList.remove('bg-indigo-800', 'text-white');
  btnPdfSizeWork.classList.add('text-indigo-200');

  btnMeasureConverter.classList.remove('bg-indigo-800', 'text-white');
  btnMeasureConverter.classList.add('text-indigo-200');

  if (btnHrCalculator) {
    btnHrCalculator.classList.remove('bg-indigo-800', 'text-white');
    btnHrCalculator.classList.add('text-indigo-200');
  }

  btnTimeConverter.classList.remove('bg-indigo-800', 'text-white');
  btnTimeConverter.classList.add('text-indigo-200');

  btnRailwayConverter.classList.remove('bg-indigo-800', 'text-white');
  btnRailwayConverter.classList.add('text-indigo-200');

  // Activate selected view
  if (view === 'file') {
    fileConverterView.classList.remove('hidden');
    btnFileConverter.classList.replace('text-indigo-200', 'text-white');
    btnFileConverter.classList.add('bg-indigo-800');
  } else if (view === 'pdf') {
    pdfConverterView.classList.remove('hidden');
    btnPdfSizeWork.classList.replace('text-indigo-200', 'text-white');
    btnPdfSizeWork.classList.add('bg-indigo-800');
    // Keep PDF dropdown open so the active sub-item stays visible
    pdfMenuDropdown.classList.remove('hidden');
    pdfMenuDropdown.classList.add('flex');
    pdfMenuIcon.classList.add('rotate-180');
    pdfMenuOpen = true;
  } else if (view === 'measure') {
    measureConverterView.classList.remove('hidden');
    btnMeasureConverter.classList.replace('text-indigo-200', 'text-white');
    btnMeasureConverter.classList.add('bg-indigo-800');
  } else if (view === 'age') {
    ageCalculatorPanel.classList.remove('hidden');
    bmiCalculatorPanel.classList.add('hidden');
    if (btnHrCalculator) {
      btnHrCalculator.classList.remove('bg-indigo-800', 'text-white');
      btnHrCalculator.classList.add('text-indigo-200');
    }
  } else if (view === 'hr') {
    hrCalculatorView.classList.remove('hidden');
    if (btnHrCalculator) {
      btnHrCalculator.classList.replace('text-indigo-200', 'text-white');
      btnHrCalculator.classList.add('bg-indigo-800');
    }
  } else if (view === 'time') {
    timeConverterView.classList.remove('hidden');
    btnTimeConverter.classList.replace('text-indigo-200', 'text-white');
    btnTimeConverter.classList.add('bg-indigo-800');
  } else if (view === 'railway') {
    railwayConverterView.classList.remove('hidden');
    btnRailwayConverter.classList.replace('text-indigo-200', 'text-white');
    btnRailwayConverter.classList.add('bg-indigo-800');
  }
}

// ----------------------------------------------------
// SIDEBAR CLICK LISTENERS
// ----------------------------------------------------
btnFileConverter.addEventListener('click', () => setActiveView('file'));
btnPdfSizeWork.addEventListener('click', () => setActiveView('pdf'));
btnMeasureConverter.addEventListener('click', () => setActiveView('measure'));

if (btnHrCalculator) {
  const hrCalcDropdown = document.getElementById('hr-calc-dropdown') as HTMLDivElement | null;
  btnHrCalculator.addEventListener('click', () => {
    if (hrCalcDropdown) {
      const isHidden = hrCalcDropdown.classList.contains('hidden');
      if (isHidden) {
        hrCalcDropdown.classList.remove('hidden');
        hrCalcDropdown.classList.add('flex');
      } else {
        hrCalcDropdown.classList.add('hidden');
        hrCalcDropdown.classList.remove('flex');
      }
    }
  });

  const btnBmi = document.getElementById('btn-bmi') as HTMLButtonElement | null;
  const btnAge = document.getElementById('btn-age') as HTMLButtonElement | null;

  btnBmi?.addEventListener('click', () => {
    setActiveView('hr');
    showHrPanel('bmi');
  });

  btnAge?.addEventListener('click', () => {
    setActiveView('age');
    if (hrCalcDropdown) {
      hrCalcDropdown.classList.add('hidden');
      hrCalcDropdown.classList.remove('flex');
    }
  });
}

btnTimeConverter.addEventListener('click', () => setActiveView('time'));
btnRailwayConverter.addEventListener('click', () => setActiveView('railway'));

// ----------------------------------------------------
// HR PANEL (BMI / AGE) INTERNAL SWITCHER
// ----------------------------------------------------
const hrModeButtons = document.querySelectorAll('[data-hr-mode]') as NodeListOf<HTMLButtonElement>;
const hrToggleButtons = document.querySelectorAll('[data-hr-mode-toggle]') as NodeListOf<HTMLButtonElement>;
const ageCalculatorPanel = document.getElementById('age-calculator-view') as HTMLDivElement;
const bmiCalculatorPanel = document.getElementById('hr-calculator-view') as HTMLDivElement;

function showHrPanel(mode: 'age' | 'bmi') {
  const isAge = mode === 'age';
  ageCalculatorPanel.classList.toggle('hidden', !isAge);
  bmiCalculatorPanel.classList.toggle('hidden', isAge);

  hrToggleButtons.forEach((btn) => {
    const active = btn.dataset.hrModeToggle === mode;
    btn.classList.toggle('bg-violet-600', active);
    btn.classList.toggle('text-white', active);
    btn.classList.toggle('bg-violet-100', !active);
    btn.classList.toggle('text-violet-700', !active);
  });

  hrModeButtons.forEach((btn) => {
    const active = btn.dataset.hrMode === mode;
    btn.classList.toggle('bg-indigo-800', active);
    btn.classList.toggle('text-white', active);
    btn.classList.toggle('text-indigo-200', !active);
  });
}

hrModeButtons.forEach((btn) => {
  btn.addEventListener('click', () => {
    showHrPanel(btn.dataset.hrMode as 'age' | 'bmi');
    setActiveView('hr');
    if (hrMenuDropdown && hrMenuIcon) {
      hrMenuDropdown.classList.add('hidden');
      hrMenuDropdown.classList.remove('flex');
      hrMenuIcon.classList.remove('rotate-180');
    }
    hrMenuOpen = false;
  });
});

hrToggleButtons.forEach((btn) => {
  btn.addEventListener('click', () => {
    showHrPanel(btn.dataset.hrModeToggle as 'age' | 'bmi');
  });
});

setActiveView('file'); // Default view on load

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
// PDF SIZE CONVERTER LOGIC
// ----------------------------------------------------
const pdfForm = document.getElementById('pdf-form') as HTMLFormElement;
const pdfFileInput = document.getElementById('pdf-file-input') as HTMLInputElement;
const pdfConversionType = document.getElementById('pdf-conversion-type') as HTMLSelectElement;
const pdfTargetSize = document.getElementById('pdf-target-size') as HTMLInputElement;
const pdfDataType = document.getElementById('pdf-data-type') as HTMLSelectElement;
const pdfStatusMessage = document.getElementById('pdf-status-message') as HTMLParagraphElement;

pdfForm.addEventListener('submit', async (e) => {
  e.preventDefault();

  pdfStatusMessage.classList.remove('hidden', 'text-red-600', 'text-green-600');
  pdfStatusMessage.classList.add('text-gray-500');

  if (!pdfFileInput.files || pdfFileInput.files.length === 0) {
    pdfStatusMessage.textContent = 'Please select a PDF file first.';
    pdfStatusMessage.classList.add('text-red-600');
    return;
  }

  const file = pdfFileInput.files[0];
  const formData = new FormData();
  formData.append('file', file);
  formData.append('conversionType', pdfConversionType.value);
  formData.append('dataType', pdfDataType.value);
  formData.append('targetSize', pdfTargetSize.value);

  try {
    pdfStatusMessage.textContent = `Converting PDF...`;

    const isLocal = window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1';
    const apiUrl = isLocal
      ? 'http://localhost:8080/convert-pdf-size'
      : 'https://utils.api.srilakshmiretail.in/convert-pdf-size';

    const response = await fetch(apiUrl, {
      method: 'POST',
      body: formData,
    });

    if (!response.ok) {
      const errText = await response.text();
      throw new Error(errText || `Server error: ${response.status}`);
    }

    // Try to get the filename from the Content-Disposition header
    const contentDisposition = response.headers.get('Content-Disposition');
    let filename = 'converted.pdf';
    if (contentDisposition && contentDisposition.includes('filename=')) {
      filename = contentDisposition.split('filename=')[1].replace(/"/g, '');
    }

    const blob = await response.blob();
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.style.display = 'none';
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    window.URL.revokeObjectURL(url);
    document.body.removeChild(a);

    pdfStatusMessage.textContent = `Success! PDF converted and downloading.`;
    pdfStatusMessage.classList.replace('text-gray-500', 'text-green-600');
  } catch (error) {
    console.error('PDF Conversion error:', error);
    pdfStatusMessage.textContent = `Error: ${error instanceof Error ? error.message : 'Unknown error occurred'}`;
    pdfStatusMessage.classList.replace('text-gray-500', 'text-red-600');
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
  speed: ["meters_per_second", "kilometers_per_hour", "miles_per_hour", "feet_per_second", "knots"],
  data: ["bytes", "kilobytes", "megabytes", "gigabytes", "terabytes", "petabytes", "bits"],
  numeral: ["binary", "octal", "decimal", "hexadecimal"]
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
    const label = unit.split('_').map(w => w.charAt(0).toUpperCase() + w.slice(1)).join(' ');
    measureFrom.add(new Option(label, unit));
    measureTo.add(new Option(label, unit));
  });

  if (units.length > 1) {
    measureTo.selectedIndex = 1;
  }

  if (category === 'numeral') {
    measureValue.type = 'text';
    measureValue.placeholder = 'e.g., 1011 or FF';
  } else {
    measureValue.type = 'number';
    measureValue.placeholder = 'Enter value to convert';
  }
}

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

  const isNumeral = measureCategory.value === 'numeral';

  const payload = isNumeral ? {
    fromBase: measureFrom.value,
    toBase: measureTo.value,
    value: measureValue.value
  } : {
    category: measureCategory.value,
    fromUnit: measureFrom.value,
    toUnit: measureTo.value,
    value: parseFloat(measureValue.value)
  };

  try {
    const isLocal = window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1';

    let apiPath = isNumeral ? '/convert-numeral' : '/convert-measurement';
    const apiUrl = isLocal
      ? `http://localhost:8080${apiPath}`
      : `https://utils.api.srilakshmiretail.in${apiPath}`;

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

    let formattedResult;
    if (isNumeral) {
      formattedResult = data.result;
    } else {
      formattedResult = Number.isInteger(data.result) ? data.result : Number(data.result.toFixed(6));
    }

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

  const dateValue = timeDateInput.value;
  if (!dateValue) return;

  const dateParts = dateValue.split('-');

  let hour = parseInt(timeHourInput.value, 10);
  const minute = parseInt(timeMinuteInput.value, 10);
  const ampm = timeAmpmInput.value;

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

// ----------------------------------------------------
// RAILWAY CONVERTER LOGIC
// ----------------------------------------------------
const rw12Hour = document.getElementById('railway-12-hour') as HTMLInputElement;
const rw12Min = document.getElementById('railway-12-min') as HTMLInputElement;
const rw12Ampm = document.getElementById('railway-12-ampm') as HTMLSelectElement;

const rw24Hour = document.getElementById('railway-24-hour') as HTMLInputElement;
const rw24Min = document.getElementById('railway-24-min') as HTMLInputElement;

const railwayStatusMessage = document.getElementById('railway-status-message') as HTMLParagraphElement;

let isUpdatingRailway = false;

async function sync12to24() {
  if (isUpdatingRailway) return;

  const h = parseInt(rw12Hour.value, 10);
  const m = parseInt(rw12Min.value, 10);
  if (isNaN(h) || isNaN(m)) return;

  isUpdatingRailway = true;

  const payload = {
    direction: "12to24",
    hour: h,
    minute: m,
    ampm: rw12Ampm.value
  };

  try {
    const isLocal = window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1';
    const apiUrl = isLocal
      ? 'http://localhost:8080/convert-railway'
      : 'https://utils.api.srilakshmiretail.in/convert-railway';

    const res = await fetch(apiUrl, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });

    if (res.ok) {
      const data = await res.json();
      rw24Hour.value = data.hour_24.toString().padStart(2, '0');
      rw24Min.value = data.minute.toString().padStart(2, '0');
      railwayStatusMessage.classList.add('hidden');
    } else {
      throw new Error(await res.text());
    }
  } catch (err) {
    console.error("Railway convert error", err);
    railwayStatusMessage.textContent = `Error: ${err instanceof Error ? err.message : 'Unknown'}`;
    railwayStatusMessage.classList.remove('hidden');
  } finally {
    isUpdatingRailway = false;
  }
}

async function sync24to12() {
  if (isUpdatingRailway) return;

  const h = parseInt(rw24Hour.value, 10);
  const m = parseInt(rw24Min.value, 10);
  if (isNaN(h) || isNaN(m)) return;

  isUpdatingRailway = true;

  const payload = {
    direction: "24to12",
    hour: h,
    minute: m
  };

  try {
    const isLocal = window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1';
    const apiUrl = isLocal
      ? 'http://localhost:8080/convert-railway'
      : 'https://utils.api.srilakshmiretail.in/convert-railway';

    const res = await fetch(apiUrl, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });

    if (res.ok) {
      const data = await res.json();
      rw12Hour.value = data.hour_12.toString().padStart(2, '0');
      rw12Min.value = data.minute.toString().padStart(2, '0');
      rw12Ampm.value = data.ampm;
      railwayStatusMessage.classList.add('hidden');
    } else {
      throw new Error(await res.text());
    }
  } catch (err) {
    console.error("Railway convert error", err);
    railwayStatusMessage.textContent = `Error: ${err instanceof Error ? err.message : 'Unknown'}`;
    railwayStatusMessage.classList.remove('hidden');
  } finally {
    isUpdatingRailway = false;
  }
}

['input', 'change'].forEach(evt => {
  rw12Hour.addEventListener(evt, sync12to24);
  rw12Min.addEventListener(evt, sync12to24);
  rw12Ampm.addEventListener(evt, sync12to24);

  rw24Hour.addEventListener(evt, sync24to12);
  rw24Min.addEventListener(evt, sync24to12);
});

rw12Hour.value = "12";
rw12Min.value = "00";
rw12Ampm.value = "PM";
sync12to24();

// ----------------------------------------------------
// AGE CALCULATOR LOGIC
// ----------------------------------------------------
const ageForm = document.getElementById('age-form') as HTMLFormElement;
const ageDob = document.getElementById('age-dob') as HTMLInputElement;
const ageToday = document.getElementById('age-today') as HTMLInputElement;
const ageResultBox = document.getElementById('age-result-box') as HTMLDivElement;
const ageResultText = document.getElementById('age-result-text') as HTMLParagraphElement;
const ageNextBirthday = document.getElementById('age-next-birthday') as HTMLParagraphElement;
const ageSummaryText = document.getElementById('age-summary-text') as HTMLParagraphElement;
const ageStatusMessage = document.getElementById('age-status-message') as HTMLParagraphElement;

if (ageForm) {
  const today = new Date();
  const isoToday = today.toISOString().split('T')[0];
  ageToday.value = isoToday;

  ageForm.addEventListener('submit', async (e) => {
    e.preventDefault();

    ageResultBox.classList.add('hidden');
    ageStatusMessage.classList.add('hidden');

    const dob = ageDob.value;
    const todayDate = ageToday.value;

    if (!dob || !todayDate) {
      ageStatusMessage.textContent = 'Please enter both dates.';
      ageStatusMessage.classList.remove('hidden');
      ageStatusMessage.classList.add('text-red-600');
      return;
    }

    try {
      const isLocal = window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1';
      const apiUrl = isLocal
        ? 'http://localhost:8080/calculate-age'
        : 'https://utils.api.srilakshmiretail.in/calculate-age';

      const response = await fetch(apiUrl, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ dob, today: todayDate })
      });

      if (!response.ok) {
        const errText = await response.text();
        throw new Error(errText || `Server error: ${response.status}`);
      }

      const data = await response.json();
      const age = data.age;
      const nextBirthday = data.nextBirthday;
      const summary = data.summary;

      ageResultText.textContent = `${age.years} Years, ${age.months} Months, ${age.days} Days`;
      ageNextBirthday.textContent = `${nextBirthday.dayOfWeek} - ${nextBirthday.date} (${nextBirthday.month}/${nextBirthday.day})`;
      ageSummaryText.textContent = `${summary.years}y, ${summary.months}m, ${summary.weeks}w, ${summary.days}d, ${summary.hours}h, ${summary.minutes}m`;

      ageResultBox.classList.remove('hidden');
    } catch (error) {
      console.error('Age calculation error:', error);
      ageStatusMessage.textContent = `Error: ${error instanceof Error ? error.message : 'Unknown error occurred'}`;
      ageStatusMessage.classList.remove('hidden');
      ageStatusMessage.classList.add('text-red-600');
    }
  });
}

// ----------------------------------------------------
// BMI CALCULATOR LOGIC
// ----------------------------------------------------
const bmiForm = document.getElementById('bmi-form') as HTMLFormElement;
const bmiWeight = document.getElementById('bmi-weight') as HTMLInputElement;
const bmiWeightUnit = document.getElementById('bmi-weight-unit') as HTMLSelectElement;
const bmiHeight = document.getElementById('bmi-height') as HTMLInputElement;
const bmiHeightUnit = document.getElementById('bmi-height-unit') as HTMLSelectElement;
const bmiResultSection = document.getElementById('bmi-result-section') as HTMLDivElement;
const bmiResultValue = document.getElementById('bmi-result-value') as HTMLSpanElement;
const bmiResultCategory = document.getElementById('bmi-result-category') as HTMLSpanElement;
const bmiErrorMessage = document.getElementById('bmi-error-message') as HTMLDivElement;

if (bmiForm) {
  bmiForm.addEventListener('submit', async (e) => {
    e.preventDefault();

    bmiResultSection.classList.add('hidden');
    bmiErrorMessage.classList.add('hidden');

    const payload = {
      weight: parseFloat(bmiWeight.value),
      weightUnit: bmiWeightUnit.value,
      height: parseFloat(bmiHeight.value),
      heightUnit: bmiHeightUnit.value
    };

    try {
      const isLocal = window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1';
      const apiUrl = isLocal
        ? 'http://localhost:8080/calculate-bmi'
        : 'https://utils.api.srilakshmiretail.in/calculate-bmi';

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

      bmiResultValue.textContent = data.bmi.toFixed(2);
      bmiResultCategory.textContent = data.category;

      bmiResultValue.className = 'text-4xl font-extrabold';
      if (data.category === 'Underweight') {
        bmiResultValue.classList.add('text-blue-500');
      } else if (data.category === 'Normal weight') {
        bmiResultValue.classList.add('text-green-500');
      } else if (data.category === 'Overweight') {
        bmiResultValue.classList.add('text-yellow-500');
      } else {
        bmiResultValue.classList.add('text-red-500');
      }

      bmiResultSection.classList.remove('hidden');
    } catch (error) {
      console.error('BMI calculation error:', error);
      bmiErrorMessage.textContent = `Error: ${error instanceof Error ? error.message : 'Unknown error occurred'}`;
      bmiErrorMessage.classList.remove('hidden');
    }
  });
}