const fs = require('fs');
let code = fs.readFileSync('frontend/src/main.ts', 'utf8');

const regex = /\/\/ -{52}\n\/\/ RAILWAY CONVERTER LOGIC\n\/\/ -{52}[\s\S]*/;

const newLogic = `// ----------------------------------------------------
// RAILWAY CONVERTER LOGIC
// ----------------------------------------------------
const railwayForm = document.getElementById('railway-form') as HTMLFormElement;
const railwayType = document.getElementById('railway-type') as HTMLSelectElement;

const railwayInput12h = document.getElementById('railway-input-12h') as HTMLDivElement;
const rw12Hour = document.getElementById('railway-12-hour') as HTMLSelectElement;
const rw12Min = document.getElementById('railway-12-min') as HTMLSelectElement;
const rw12Ampm = document.getElementById('railway-12-ampm') as HTMLSelectElement;

const railwayInput24h = document.getElementById('railway-input-24h') as HTMLDivElement;
const rw24Hour = document.getElementById('railway-24-hour') as HTMLSelectElement;
const rw24Min = document.getElementById('railway-24-min') as HTMLSelectElement;

const railwayResultBox = document.getElementById('railway-result-box') as HTMLDivElement;
const railwayResultText = document.getElementById('railway-result-text') as HTMLSpanElement;
const railwayStatusMessage = document.getElementById('railway-status-message') as HTMLParagraphElement;

// Populate
for (let h = 1; h <= 12; h++) {
  const hr = h.toString().padStart(2, '0');
  rw12Hour.add(new Option(hr, hr));
}
for (let h = 0; h <= 23; h++) {
  const hr = h.toString().padStart(2, '0');
  rw24Hour.add(new Option(hr, hr));
}
for (let m = 0; m <= 59; m++) {
  const min = m.toString().padStart(2, '0');
  rw12Min.add(new Option(min, min));
  rw24Min.add(new Option(min, min));
}

rw12Hour.value = "12";
rw12Min.value = "00";
rw12Ampm.value = "PM";

rw24Hour.value = "12";
rw24Min.value = "00";

railwayType.addEventListener('change', () => {
  if (railwayType.value === "12to24") {
    railwayInput12h.classList.remove('hidden');
    railwayInput24h.classList.add('hidden');
  } else {
    railwayInput12h.classList.add('hidden');
    railwayInput24h.classList.remove('hidden');
  }
  railwayResultBox.classList.add('hidden');
  railwayStatusMessage.classList.add('hidden');
});

railwayForm.addEventListener('submit', async (e) => {
  e.preventDefault();
  
  railwayStatusMessage.classList.remove('hidden', 'text-red-600', 'text-green-600');
  railwayStatusMessage.classList.add('text-gray-500');
  railwayStatusMessage.textContent = 'Converting...';
  railwayResultBox.classList.add('hidden');

  let payload: any = {
    direction: railwayType.value
  };

  if (railwayType.value === "12to24") {
    payload.hour = parseInt(rw12Hour.value, 10);
    payload.minute = parseInt(rw12Min.value, 10);
    payload.ampm = rw12Ampm.value;
  } else {
    payload.hour = parseInt(rw24Hour.value, 10);
    payload.minute = parseInt(rw24Min.value, 10);
  }

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

    if (!res.ok) {
      throw new Error(await res.text() || 'Server error');
    }

    const data = await res.json();
    
    if (payload.direction === "12to24") {
      const h24 = data.hour_24.toString().padStart(2, '0');
      const m24 = data.minute.toString().padStart(2, '0');
      railwayResultText.textContent = \`\${h24}:\${m24} Railway Time\`;
    } else {
      const h12 = data.hour_12.toString().padStart(2, '0');
      const m12 = data.minute.toString().padStart(2, '0');
      railwayResultText.textContent = \`\${h12}:\${m12} \${data.ampm}\`;
    }
    
    railwayResultBox.classList.remove('hidden');
    railwayStatusMessage.classList.add('hidden');
  } catch (err) {
    console.error("Railway convert error", err);
    railwayStatusMessage.textContent = \`Error: \${err instanceof Error ? err.message : 'Unknown error'}\`;
    railwayStatusMessage.classList.replace('text-gray-500', 'text-red-600');
  }
});
`;

code = code.replace(regex, newLogic);
fs.writeFileSync('frontend/src/main.ts', code);
