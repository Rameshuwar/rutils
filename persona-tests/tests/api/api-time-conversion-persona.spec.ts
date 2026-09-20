import { test, expect } from '@playwright/test';

const API_URL = 'http://localhost:8080';

test.describe('Backend API Time & Railway Conversion Persona', () => {

  // TIMEZONE TESTS
  test('Persona: Successful Timezone Conversion (Standard)', async ({ request }) => {
    await test.step('Why we use this test: We need to ensure standard timezone conversions work reliably.', async () => {});
    
    let response;
    
    await test.step('What we use: We request the conversion of a specific time in Asia/Kolkata to America/New_York.', async () => {
      response = await request.post(`${API_URL}/convert-time`, {
        data: {
          year: 2024,
          month: 1,
          day: 1,
          hour: 12,
          minute: 0,
          second: 0,
          source_tz: 'Asia/Kolkata',
          dest_tz: 'America/New_York'
        }
      });
    });

    await test.step('What we expected: We expect a 200 OK status.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: We read the returned JSON and verify the output time.', async () => {
      const jsonResponse = await response.json();
      expect(jsonResponse.resolution).toBe('normal');
      expect(jsonResponse.dest_zone_name).toBe('EST'); // Jan 1st is standard time in NY (EST)
      expect(jsonResponse.dest_time_local).toContain('01:30:00'); // 12:00 IST = 06:30 UTC = 01:30 EST
    });

    await test.step('Why it got this output: The backend correctly computed the time difference between Kolkata (+5:30) and New York (-5:00) during standard time.', async () => {});
  });

  test('Persona: Timezone Conversion with Next Day', async ({ request }) => {
    await test.step('Why we use this test: We want to check if date boundary crossing is accurately reported.', async () => {});
    
    let response;
    
    await test.step('What we use: We convert a late evening time in New York to Tokyo.', async () => {
      response = await request.post(`${API_URL}/convert-time`, {
        data: {
          year: 2024,
          month: 1,
          day: 1,
          hour: 23,
          minute: 0,
          second: 0,
          source_tz: 'America/New_York',
          dest_tz: 'Asia/Tokyo'
        }
      });
    });

    await test.step('What we expected: We expect a 200 OK status.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: We read the returned JSON and ensure is_next_day is true.', async () => {
      const jsonResponse = await response.json();
      expect(jsonResponse.is_next_day).toBe(true);
      expect(jsonResponse.dest_time_local).toContain('13:00:00'); // 23:00 EST = 04:00+1d UTC = 13:00+1d JST
    });

    await test.step('Why it got this output: 11 PM in New York corresponds to 1 PM the next day in Tokyo, triggering the is_next_day flag.', async () => {});
  });

  test('Persona: Ambiguous Time Conversion', async ({ request }) => {
    await test.step('Why we use this test: Daylight saving time fallback creates ambiguous hours. We need to verify the server handles it.', async () => {});
    
    let response;
    
    await test.step('What we use: We request an ambiguous time during DST fallback in New York.', async () => {
      response = await request.post(`${API_URL}/convert-time`, {
        data: {
          year: 2023,
          month: 11,
          day: 5,
          hour: 1,
          minute: 30,
          second: 0,
          source_tz: 'America/New_York',
          dest_tz: 'UTC',
          ambiguous_policy: 'second'
        }
      });
    });

    await test.step('What we expected: We expect a 200 OK status.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: We check that resolution is ambiguous and the second occurrence is used.', async () => {
      const jsonResponse = await response.json();
      expect(jsonResponse.resolution).toBe('ambiguous');
      expect(jsonResponse.source_offset).toBe('-05:00'); // The second occurrence is in EST (-05:00)
    });

    await test.step('Why it got this output: The time 1:30 AM happens twice on Nov 5, 2023 in New York. The backend honored the ambiguous_policy of "second".', async () => {});
  });

  test('Persona: Invalid Timezone Error', async ({ request }) => {
    await test.step('Why we use this test: We want to ensure invalid timezones are rejected gracefully.', async () => {});
    
    let response;
    
    await test.step('What we use: We send a request with a fake timezone "Earth/Atlantis".', async () => {
      response = await request.post(`${API_URL}/convert-time`, {
        data: {
          year: 2024, month: 1, day: 1, hour: 12, minute: 0, second: 0,
          source_tz: 'Earth/Atlantis',
          dest_tz: 'UTC'
        }
      });
    });

    await test.step('What we expected: We expect a 400 Bad Request error.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: We verify the error message.', async () => {
      const textResponse = await response.text();
      expect(textResponse).toContain('invalid source timezone');
    });

    await test.step('Why it got this output: The standard IANA timezone database does not contain "Earth/Atlantis".', async () => {});
  });

  // RAILWAY TIME TESTS
  test('Persona: Successful Railway Conversion (12to24 PM)', async ({ request }) => {
    await test.step('Why we use this test: We need to verify that afternoon times convert correctly to 24-hour format.', async () => {});
    
    let response;
    
    await test.step('What we use: We convert 8:30 PM to 24-hour time.', async () => {
      response = await request.post(`${API_URL}/convert-railway`, {
        data: {
          direction: '12to24',
          hour: 8,
          minute: 30,
          ampm: 'PM'
        }
      });
    });

    await test.step('What we expected: We expect a 200 OK status.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: We expect 20 for the 24-hour format.', async () => {
      const jsonResponse = await response.json();
      expect(jsonResponse.hour_24).toBe(20);
      expect(jsonResponse.minute).toBe(30);
    });

    await test.step('Why it got this output: 8 PM is represented as 20 in 24-hour format by adding 12.', async () => {});
  });

  test('Persona: Successful Railway Conversion (24to12 AM)', async ({ request }) => {
    await test.step('Why we use this test: We need to verify that morning hours in 24-hour format are converted properly to 12-hour AM.', async () => {});
    
    let response;
    
    await test.step('What we use: We convert 00:15 (midnight) to 12-hour time.', async () => {
      response = await request.post(`${API_URL}/convert-railway`, {
        data: {
          direction: '24to12',
          hour: 0,
          minute: 15
        }
      });
    });

    await test.step('What we expected: We expect a 200 OK status.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: We expect 12:15 AM.', async () => {
      const jsonResponse = await response.json();
      expect(jsonResponse.hour_12).toBe(12);
      expect(jsonResponse.ampm).toBe('AM');
      expect(jsonResponse.minute).toBe(15);
    });

    await test.step('Why it got this output: Hour 0 in a 24-hour clock corresponds to 12 AM.', async () => {});
  });

  test('Persona: Invalid Railway Minute Error', async ({ request }) => {
    await test.step('Why we use this test: We want to ensure inputs that are out of bounds for minutes are rejected.', async () => {});
    
    let response;
    
    await test.step('What we use: We send a railway time conversion request with 75 minutes.', async () => {
      response = await request.post(`${API_URL}/convert-railway`, {
        data: {
          direction: '24to12',
          hour: 12,
          minute: 75
        }
      });
    });

    await test.step('What we expected: We expect a 400 Bad Request error.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: We verify the error message.', async () => {
      const textResponse = await response.text();
      expect(textResponse).toContain('minute must be between 0 and 59');
    });

    await test.step('Why it got this output: The backend strictly enforces valid minute ranges.', async () => {});
  });

  test('Persona: Invalid Direction Error', async ({ request }) => {
    await test.step('Why we use this test: We want to ensure that unknown directions are properly rejected.', async () => {});
    
    let response;
    
    await test.step('What we use: We send a railway time conversion request with an invalid direction "upwards".', async () => {
      response = await request.post(`${API_URL}/convert-railway`, {
        data: {
          direction: 'upwards',
          hour: 12,
          minute: 30
        }
      });
    });

    await test.step('What we expected: We expect a 400 Bad Request error.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: We verify the error message.', async () => {
      const textResponse = await response.text();
      expect(textResponse).toContain('direction must be');
    });

    await test.step('Why it got this output: The backend only accepts "12to24" and "24to12" for the direction parameter.', async () => {});
  });

});
