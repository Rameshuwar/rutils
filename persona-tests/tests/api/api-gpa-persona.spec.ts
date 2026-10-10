import { test, expect } from '@playwright/test';

// Configuration for API testing. The Swagger API is hosted at localhost:8080 by default.
const API_URL = 'http://localhost:8080';

test.describe('Backend API GPA / CGPA Calculator Persona', () => {

  // ================================================================
  // GROUP 1: semester_gpa
  // ================================================================

  test('Persona: Semester GPA from a course list (standard 10-point scale)', async ({ request }) => {
    await test.step('Why we use this test: The core use case — a student enters their courses with credits and grade points, and receives a credit-weighted GPA.', async () => {});

    let response;

    await test.step('What we use: POST /calculate-gpa with mode=semester_gpa and three courses.', async () => {
      response = await request.post(`${API_URL}/calculate-gpa`, {
        data: {
          mode: 'semester_gpa',
          courses: [
            { name: 'Math',      credits: 4, gradePoint: 9 },
            { name: 'Physics',   credits: 3, gradePoint: 8 },
            { name: 'Chemistry', credits: 3, gradePoint: 10 },
          ],
        },
      });
    });

    await test.step('What we expected: 200 OK.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: GPA = (4×9 + 3×8 + 3×10) / 10 = 9.0, with a 3-row breakdown and formula steps.', async () => {
      const json = await response.json();
      expect(json.mode).toBe('semester_gpa');
      expect(json.gpa).toBe(9);
      expect(json.totalCredits).toBe(10);
      expect(json.totalWeighted).toBe(90);
      expect(json.formatted).toBe('9');
      expect(Array.isArray(json.breakdown)).toBe(true);
      expect(json.breakdown).toHaveLength(3);
      expect(json.breakdown[0].weighted).toBe(36);
      expect(Array.isArray(json.steps)).toBe(true);
      expect(json.steps.length).toBeGreaterThan(0);
    });

    await test.step('Why it got this output: The backend applied GPA = Σ(credits × gradePoint) / Σ(credits).', async () => {});
  });

  test('Persona: Semester GPA with a single course', async ({ request }) => {
    await test.step('Why we use this test: Verify the simplest possible semester — one course — where GPA must equal that course\'s grade point exactly.', async () => {});

    let response;

    await test.step('What we use: POST with one course at 5 credits and 8 grade points.', async () => {
      response = await request.post(`${API_URL}/calculate-gpa`, {
        data: {
          mode: 'semester_gpa',
          courses: [{ credits: 5, gradePoint: 8 }],
        },
      });
    });

    await test.step('What we expected: 200 OK.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: GPA = 8.', async () => {
      const json = await response.json();
      expect(json.gpa).toBe(8);
      expect(json.totalCredits).toBe(5);
    });

    await test.step('Why it got this output: A single-course weighted average is trivially the grade point itself.', async () => {});
  });

  test('Persona: Semester GPA with fractional credits', async ({ request }) => {
    await test.step('Why we use this test: Some institutions award 0.5-credit or 1.5-credit courses; the weighted math must handle them.', async () => {});

    let response;

    await test.step('What we use: POST with 1.5 and 2.5 credits.', async () => {
      response = await request.post(`${API_URL}/calculate-gpa`, {
        data: {
          mode: 'semester_gpa',
          courses: [
            { credits: 1.5, gradePoint: 9 },
            { credits: 2.5, gradePoint: 7 },
          ],
        },
      });
    });

    await test.step('What we expected: 200 OK.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: GPA = (1.5×9 + 2.5×7) / 4 = 31/4 = 7.75.', async () => {
      const json = await response.json();
      expect(json.gpa).toBeCloseTo(7.75, 2);
      expect(json.totalCredits).toBe(4);
    });

    await test.step('Why it got this output: Fractional credits are accumulated exactly and the weighted sum is divided by the total.', async () => {});
  });

  test('Persona: Semester GPA with perfect score', async ({ request }) => {
    await test.step('Why we use this test: All O grades (10 points) must yield a GPA of 10.', async () => {});

    let response;

    await test.step('What we use: POST with two 4-credit courses at grade point 10.', async () => {
      response = await request.post(`${API_URL}/calculate-gpa`, {
        data: {
          mode: 'semester_gpa',
          courses: [
            { credits: 4, gradePoint: 10 },
            { credits: 4, gradePoint: 10 },
          ],
        },
      });
    });

    await test.step('What we expected: 200 OK.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: GPA = 10.', async () => {
      const json = await response.json();
      expect(json.gpa).toBe(10);
    });

    await test.step('Why it got this output: Every weighted contribution equals credits × 10, and the divisor is the same sum.', async () => {});
  });

  test('Persona: Error - semester GPA without courses', async ({ request }) => {
    await test.step('Why we use this test: An empty course list makes the weighted average undefined; the backend must reject with a helpful message.', async () => {});

    let response;

    await test.step('What we use: POST with mode=semester_gpa and no courses field.', async () => {
      response = await request.post(`${API_URL}/calculate-gpa`, {
        data: { mode: 'semester_gpa' },
      });
    });

    await test.step('What we expected: 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The error mentions courses are required.', async () => {
      const txt = await response.text();
      expect(txt).toContain('courses are required');
    });

    await test.step('Why it got this output: The dispatcher validates the payload before touching the math.', async () => {});
  });

  test('Persona: Error - semester GPA with zero credits', async ({ request }) => {
    await test.step('Why we use this test: A zero-credit course is meaningless and would silently distort the divisor; reject it.', async () => {});

    let response;

    await test.step('What we use: POST with a single course at credits=0.', async () => {
      response = await request.post(`${API_URL}/calculate-gpa`, {
        data: {
          mode: 'semester_gpa',
          courses: [{ credits: 0, gradePoint: 9 }],
        },
      });
    });

    await test.step('What we expected: 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The error mentions credits must be greater than zero.', async () => {
      const txt = await response.text();
      expect(txt).toContain('credits must be greater than zero');
    });

    await test.step('Why it got this output: The validator guards every course entry before computing.', async () => {});
  });

  test('Persona: Error - semester GPA with grade point above 10', async ({ request }) => {
    await test.step('Why we use this test: The 10-point scale is the hard ceiling for this mode; anything above is a client bug.', async () => {});

    let response;

    await test.step('What we use: POST with a course at gradePoint=11.', async () => {
      response = await request.post(`${API_URL}/calculate-gpa`, {
        data: {
          mode: 'semester_gpa',
          courses: [{ credits: 3, gradePoint: 11 }],
        },
      });
    });

    await test.step('What we expected: 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The error mentions gradePoint cannot exceed 10.', async () => {
      const txt = await response.text();
      expect(txt).toContain('gradePoint cannot exceed 10');
    });

    await test.step('Why it got this output: The per-course validator enforces the 10-point upper bound.', async () => {});
  });

  // ================================================================
  // GROUP 2: cumulative_cgpa
  // ================================================================

  test('Persona: Cumulative CGPA across three semesters', async ({ request }) => {
    await test.step('Why we use this test: The credit-weighted CGPA across multiple semesters is the primary use case for a graduating student.', async () => {});

    let response;

    await test.step('What we use: POST with three 20-credit semesters at GPA 8, 9, and 9.5.', async () => {
      response = await request.post(`${API_URL}/calculate-gpa`, {
        data: {
          mode: 'cumulative_cgpa',
          semesters: [
            { name: 'Sem 1', credits: 20, gpa: 8 },
            { name: 'Sem 2', credits: 20, gpa: 9 },
            { name: 'Sem 3', credits: 20, gpa: 9.5 },
          ],
        },
      });
    });

    await test.step('What we expected: 200 OK.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: CGPA = (160 + 180 + 190) / 60 = 8.83, with a 3-row semester breakdown.', async () => {
      const json = await response.json();
      expect(json.mode).toBe('cumulative_cgpa');
      expect(json.cgpa).toBeCloseTo(8.83, 2);
      expect(json.totalCredits).toBe(60);
      expect(Array.isArray(json.semesters)).toBe(true);
      expect(json.semesters).toHaveLength(3);
    });

    await test.step('Why it got this output: The backend applied CGPA = Σ(credits × gpa) / Σ(credits).', async () => {});
  });

  test('Persona: Cumulative CGPA with unequal credits', async ({ request }) => {
    await test.step('Why we use this test: Real transcripts have semesters with different total credits; weighting must reflect that.', async () => {});

    let response;

    await test.step('What we use: POST with a 24-credit semester at 9 and an 18-credit semester at 7.', async () => {
      response = await request.post(`${API_URL}/calculate-gpa`, {
        data: {
          mode: 'cumulative_cgpa',
          semesters: [
            { credits: 24, gpa: 9 },
            { credits: 18, gpa: 7 },
          ],
        },
      });
    });

    await test.step('What we expected: 200 OK.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: CGPA = (216 + 126) / 42 = 8.14.', async () => {
      const json = await response.json();
      expect(json.cgpa).toBeCloseTo(8.14, 2);
      expect(json.totalCredits).toBe(42);
    });

    await test.step('Why it got this output: The larger semester carries proportionally more weight.', async () => {});
  });

  test('Persona: Error - cumulative CGPA without semesters', async ({ request }) => {
    await test.step('Why we use this test: An empty semester list is undefined; reject it with a clear message.', async () => {});

    let response;

    await test.step('What we use: POST with mode=cumulative_cgpa and no semesters field.', async () => {
      response = await request.post(`${API_URL}/calculate-gpa`, {
        data: { mode: 'cumulative_cgpa' },
      });
    });

    await test.step('What we expected: 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The error mentions semesters are required.', async () => {
      const txt = await response.text();
      expect(txt).toContain('semesters are required');
    });

    await test.step('Why it got this output: The dispatcher validates the payload before touching the math.', async () => {});
  });

  test('Persona: Error - cumulative CGPA with a GPA above 10', async ({ request }) => {
    await test.step('Why we use this test: Individual semester GPAs must respect the 10-point ceiling even in cumulative mode.', async () => {});

    let response;

    await test.step('What we use: POST with a semester at gpa=11.', async () => {
      response = await request.post(`${API_URL}/calculate-gpa`, {
        data: {
          mode: 'cumulative_cgpa',
          semesters: [{ credits: 20, gpa: 11 }],
        },
      });
    });

    await test.step('What we expected: 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The error mentions gpa cannot exceed 10.', async () => {
      const txt = await response.text();
      expect(txt).toContain('gpa cannot exceed 10');
    });

    await test.step('Why it got this output: The per-semester validator enforces the same 10-point bound used everywhere else.', async () => {});
  });

  // ================================================================
  // GROUP 3: cgpa_to_percentage
  // ================================================================

  test('Persona: CGPA to percentage using the CBSE 9.5 factor (default)', async ({ request }) => {
    await test.step('Why we use this test: The most common conversion in India — a CGPA of 8.0 must become 76% under the CBSE convention.', async () => {});

    let response;

    await test.step('What we use: POST with mode=cgpa_to_percentage, cgpa=8.0, no multiplier (defaults to 9.5).', async () => {
      response = await request.post(`${API_URL}/calculate-gpa`, {
        data: { mode: 'cgpa_to_percentage', cgpa: 8.0 },
      });
    });

    await test.step('What we expected: 200 OK.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: percentage = 76, multiplier echoed as 9.5, formatted = "76%".', async () => {
      const json = await response.json();
      expect(json.mode).toBe('cgpa_to_percentage');
      expect(json.percentage).toBe(76);
      expect(json.multiplier).toBe(9.5);
      expect(json.formatted).toBe('76%');
      expect(Array.isArray(json.steps)).toBe(true);
    });

    await test.step('Why it got this output: The default multiplier is 9.5 per the CBSE / AICTE convention.', async () => {});
  });

  test('Persona: CGPA to percentage using the VTU 10.0 factor', async ({ request }) => {
    await test.step('Why we use this test: Many universities (VTU, KTU) use a 10.0 multiplier instead of 9.5; the endpoint must honour the override.', async () => {});

    let response;

    await test.step('What we use: POST with cgpa=8.4 and multiplier=10.', async () => {
      response = await request.post(`${API_URL}/calculate-gpa`, {
        data: { mode: 'cgpa_to_percentage', cgpa: 8.4, multiplier: 10 },
      });
    });

    await test.step('What we expected: 200 OK.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: percentage = 84.', async () => {
      const json = await response.json();
      expect(json.percentage).toBe(84);
      expect(json.multiplier).toBe(10);
    });

    await test.step('Why it got this output: The multiplier is read directly from the request body.', async () => {});
  });

  test('Persona: CGPA to percentage using the Mumbai 7.25 factor', async ({ request }) => {
    await test.step('Why we use this test: Mumbai University uses a 7.25 multiplier — a good third data point to prove the multiplier is truly configurable.', async () => {});

    let response;

    await test.step('What we use: POST with cgpa=8.0 and multiplier=7.25.', async () => {
      response = await request.post(`${API_URL}/calculate-gpa`, {
        data: { mode: 'cgpa_to_percentage', cgpa: 8.0, multiplier: 7.25 },
      });
    });

    await test.step('What we expected: 200 OK.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: percentage = 58.', async () => {
      const json = await response.json();
      expect(json.percentage).toBe(58);
    });

    await test.step('Why it got this output: 8.0 × 7.25 = 58.0, rounded to two decimals.', async () => {});
  });

  test('Persona: Error - CGPA to percentage with negative CGPA', async ({ request }) => {
    await test.step('Why we use this test: A negative CGPA is nonsensical; reject it.', async () => {});

    let response;

    await test.step('What we use: POST with cgpa=-1.', async () => {
      response = await request.post(`${API_URL}/calculate-gpa`, {
        data: { mode: 'cgpa_to_percentage', cgpa: -1 },
      });
    });

    await test.step('What we expected: 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The error mentions cgpa cannot be negative.', async () => {
      const txt = await response.text();
      expect(txt).toContain('cgpa cannot be negative');
    });

    await test.step('Why it got this output: The validator guards the input before the arithmetic.', async () => {});
  });

  test('Persona: Error - CGPA to percentage with CGPA above 10', async ({ request }) => {
    await test.step('Why we use this test: The 10-point scale is the hard ceiling; a 10.5 CGPA is a client error.', async () => {});

    let response;

    await test.step('What we use: POST with cgpa=11.', async () => {
      response = await request.post(`${API_URL}/calculate-gpa`, {
        data: { mode: 'cgpa_to_percentage', cgpa: 11 },
      });
    });

    await test.step('What we expected: 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The error mentions cgpa cannot exceed 10.', async () => {
      const txt = await response.text();
      expect(txt).toContain('cgpa cannot exceed 10');
    });

    await test.step('Why it got this output: The validator enforces the upper bound before multiplying.', async () => {});
  });

  // ================================================================
  // GROUP 4: percentage_to_cgpa
  // ================================================================

  test('Persona: Percentage to CGPA (default 9.5)', async ({ request }) => {
    await test.step('Why we use this test: The inverse of the CBSE conversion — 76% must map back to 8.0 CGPA.', async () => {});

    let response;

    await test.step('What we use: POST with mode=percentage_to_cgpa and percentage=76.', async () => {
      response = await request.post(`${API_URL}/calculate-gpa`, {
        data: { mode: 'percentage_to_cgpa', percentage: 76 },
      });
    });

    await test.step('What we expected: 200 OK.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: cgpa = 8.0.', async () => {
      const json = await response.json();
      expect(json.cgpa).toBe(8.0);
      expect(json.multiplier).toBe(9.5);
    });

    await test.step('Why it got this output: The backend divided the percentage by the default multiplier of 9.5.', async () => {});
  });

  test('Persona: Percentage to CGPA with a custom multiplier', async ({ request }) => {
    await test.step('Why we use this test: With a 10.0 multiplier, 84% must map to 8.4 CGPA.', async () => {});

    let response;

    await test.step('What we use: POST with percentage=84 and multiplier=10.', async () => {
      response = await request.post(`${API_URL}/calculate-gpa`, {
        data: { mode: 'percentage_to_cgpa', percentage: 84, multiplier: 10 },
      });
    });

    await test.step('What we expected: 200 OK.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: cgpa = 8.4.', async () => {
      const json = await response.json();
      expect(json.cgpa).toBe(8.4);
    });

    await test.step('Why it got this output: The custom multiplier is read from the request and applied as the divisor.', async () => {});
  });

  test('Persona: Error - percentage to CGPA with percentage above 100', async ({ request }) => {
    await test.step('Why we use this test: A percentage above 100 is not a valid mark; reject it.', async () => {});

    let response;

    await test.step('What we use: POST with percentage=101.', async () => {
      response = await request.post(`${API_URL}/calculate-gpa`, {
        data: { mode: 'percentage_to_cgpa', percentage: 101 },
      });
    });

    await test.step('What we expected: 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The error mentions percentage cannot exceed 100.', async () => {
      const txt = await response.text();
      expect(txt).toContain('percentage cannot exceed 100');
    });

    await test.step('Why it got this output: The validator caps the input at 100 before the division.', async () => {});
  });

  // ================================================================
  // GROUP 5: grade_to_point
  // ================================================================

  test('Persona: Grade to point on the 10-point scale', async ({ request }) => {
    await test.step('Why we use this test: A student types "A" on a 10-point scale and expects 9 points back.', async () => {});

    let response;

    await test.step('What we use: POST with mode=grade_to_point, grade="A", gradingScale="10".', async () => {
      response = await request.post(`${API_URL}/calculate-gpa`, {
        data: { mode: 'grade_to_point', grade: 'A', gradingScale: '10' },
      });
    });

    await test.step('What we expected: 200 OK.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: gradePoint = 9.', async () => {
      const json = await response.json();
      expect(json.gradePoint).toBe(9);
      expect(json.extra.scale).toBe('10');
    });

    await test.step('Why it got this output: The backend looked up the 10-point table and found the mapping A → 9.', async () => {});
  });

  test('Persona: Grade to point on the 4-point scale (US convention)', async ({ request }) => {
    await test.step('Why we use this test: A US university transcript uses A- = 3.7 on a 4-point scale; the endpoint must produce that value.', async () => {});

    let response;

    await test.step('What we use: POST with grade="A-" and gradingScale="4".', async () => {
      response = await request.post(`${API_URL}/calculate-gpa`, {
        data: { mode: 'grade_to_point', grade: 'A-', gradingScale: '4' },
      });
    });

    await test.step('What we expected: 200 OK.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: gradePoint = 3.7.', async () => {
      const json = await response.json();
      expect(json.gradePoint).toBe(3.7);
    });

    await test.step('Why it got this output: The 4-point table maps A- to 3.7.', async () => {});
  });

  test('Persona: Grade to point is case-insensitive', async ({ request }) => {
    await test.step('Why we use this test: Users type lower-case letters; the endpoint must normalise them before lookup.', async () => {});

    let response;

    await test.step('What we use: POST with grade="a+" (lower case).', async () => {
      response = await request.post(`${API_URL}/calculate-gpa`, {
        data: { mode: 'grade_to_point', grade: 'a+' },
      });
    });

    await test.step('What we expected: 200 OK.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: gradePoint = 10.', async () => {
      const json = await response.json();
      expect(json.gradePoint).toBe(10);
    });

    await test.step('Why it got this output: The handler upper-cases the input before table lookup.', async () => {});
  });

  test('Persona: Error - grade to point with an unknown grade', async ({ request }) => {
    await test.step('Why we use this test: A grade letter that is not in the table must be rejected with a clear message.', async () => {});

    let response;

    await test.step('What we use: POST with grade="Z".', async () => {
      response = await request.post(`${API_URL}/calculate-gpa`, {
        data: { mode: 'grade_to_point', grade: 'Z' },
      });
    });

    await test.step('What we expected: 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The error mentions the grade is not recognised.', async () => {
      const txt = await response.text();
      expect(txt).toContain('not recognised');
    });

    await test.step('Why it got this output: The lookup failed and the handler surfaced the failure as a 400.', async () => {});
  });

  test('Persona: Error - grade to point with an invalid grading scale', async ({ request }) => {
    await test.step('Why we use this test: Only "10", "4", and "5" are valid scales; anything else is a client bug.', async () => {});

    let response;

    await test.step('What we use: POST with gradingScale="7".', async () => {
      response = await request.post(`${API_URL}/calculate-gpa`, {
        data: { mode: 'grade_to_point', grade: 'A', gradingScale: '7' },
      });
    });

    await test.step('What we expected: 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The error mentions gradingScale must be one of 10, 4, 5.', async () => {
      const txt = await response.text();
      expect(txt).toContain('gradingScale must be one of');
    });

    await test.step('Why it got this output: The validator rejects unknown scale identifiers before lookup.', async () => {});
  });

  // ================================================================
  // GROUP 6: target_gpa
  // ================================================================

  test('Persona: Target GPA that is achievable', async ({ request }) => {
    await test.step('Why we use this test: A student at 7.5 CGPA over 60 credits who wants to reach 8.0 over 100 total credits must be told they need 8.75 in the remaining 40 credits.', async () => {});

    let response;

    await test.step('What we use: POST with currentCgpa=7.5, completedCredits=60, targetCgpa=8.0, remainingCredits=40.', async () => {
      response = await request.post(`${API_URL}/calculate-gpa`, {
        data: {
          mode: 'target_gpa',
          currentCgpa: 7.5,
          completedCredits: 60,
          targetCgpa: 8.0,
          remainingCredits: 40,
        },
      });
    });

    await test.step('What we expected: 200 OK.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: requiredGpa = 8.75, extra.achievable = true, and no warning.', async () => {
      const json = await response.json();
      expect(json.requiredGpa).toBe(8.75);
      expect(json.totalCredits).toBe(100);
      expect(json.currentCgpa).toBe(7.5);
      expect(json.targetCgpa).toBe(8.0);
      expect(json.extra.achievable).toBe(true);
      expect(json.extra.warning).toBeUndefined();
    });

    await test.step('Why it got this output: The backend solved (8.0 × 100 − 7.5 × 60) / 40 = 350 / 40 = 8.75.', async () => {});
  });

test('Persona: Target GPA lower than current CGPA still requires nonzero performance', async ({ request }) => {
  await test.step('Why we use this test: A student at 9.0 CGPA over 60 credits who targets a lower 8.0 over 100 total credits is NOT yet done — the remaining 40 credits will pull the average down from 9.0 toward 8.0, so the student must still average 6.5 to hit the target exactly.', async () => {});

  let response;

  await test.step('What we use: POST with currentCgpa=9.0, completedCredits=60, targetCgpa=8.0, remainingCredits=40.', async () => {
    response = await request.post(`${API_URL}/calculate-gpa`, {
      data: {
        mode: 'target_gpa',
        currentCgpa: 9.0,
        completedCredits: 60,
        targetCgpa: 8.0,
        remainingCredits: 40,
      },
    });
  });

  await test.step('What we expected: 200 OK.', async () => {
    expect(response.status()).toBe(200);
  });

  await test.step('What we get: requiredGpa = 6.5 and achievable = true, no warning.', async () => {
    const json = await response.json();
    expect(json.requiredGpa).toBe(6.5);
    expect(json.extra.achievable).toBe(true);
    expect(json.extra.warning).toBeUndefined();
  });

  await test.step('Why it got this output: The backend solved (8.0 × 100 − 9.0 × 60) / 40 = 260 / 40 = 6.5. This is NOT the clamp-to-zero case — the weighted total (540) is still below the target total (800).', async () => {});
});

test('Persona: Target GPA already exceeded (required clamps to zero)', async ({ request }) => {
  await test.step('Why we use this test: When the current weighted total already exceeds the target total, the required GPA must clamp to 0 rather than going negative.', async () => {});

  let response;

  await test.step('What we use: POST with currentCgpa=9.0, completedCredits=100, targetCgpa=8.0, remainingCredits=10 — 100 credits of 9.0 already exceeds 110 credits of 8.0.', async () => {
    response = await request.post(`${API_URL}/calculate-gpa`, {
      data: {
        mode: 'target_gpa',
        currentCgpa: 9.0,
        completedCredits: 100,
        targetCgpa: 8.0,
        remainingCredits: 10,
      },
    });
  });

  await test.step('What we expected: 200 OK.', async () => {
    expect(response.status()).toBe(200);
  });

  await test.step('What we get: requiredGpa = 0 and achievable = true.', async () => {
    const json = await response.json();
    expect(json.requiredGpa).toBe(0);
    expect(json.extra.achievable).toBe(true);
  });

  await test.step('Why it got this output: currentCGPA × completedCredits = 900, targetCGPA × totalCredits = 880. 900 > 880, so the numerator is negative and the backend clamps requiredGpa to zero.', async () => {});
});

  test('Persona: Target GPA that is impossible carries a warning', async ({ request }) => {
    await test.step('Why we use this test: A student at 6.0 CGPA over 100 credits with only 10 credits left cannot reach 9.0 CGPA. The endpoint must return 200 with a warning, not a 400.', async () => {});

    let response;

    await test.step('What we use: POST with currentCgpa=6.0, completedCredits=100, targetCgpa=9.0, remainingCredits=10.', async () => {
      response = await request.post(`${API_URL}/calculate-gpa`, {
        data: {
          mode: 'target_gpa',
          currentCgpa: 6.0,
          completedCredits: 100,
          targetCgpa: 9.0,
          remainingCredits: 10,
        },
      });
    });

    await test.step('What we expected: 200 OK — the math is well-defined even if the plan is not.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: requiredGpa > 10, achievable = false, and extra.warning is present.', async () => {
      const json = await response.json();
      expect(json.requiredGpa).toBeGreaterThan(10);
      expect(json.extra.achievable).toBe(false);
      expect(typeof json.extra.warning).toBe('string');
      expect(json.extra.warning.length).toBeGreaterThan(0);
    });

    await test.step('Why it got this output: (9.0 × 110 − 6.0 × 100) / 10 = 39, which exceeds the 10-point scale. The backend flagged it rather than erroring.', async () => {});
  });

  test('Persona: Target GPA on a custom 4-point scale', async ({ request }) => {
    await test.step('Why we use this test: The same planner must work on the 4-point scale used by US universities.', async () => {});

    let response;

    await test.step('What we use: POST with currentCgpa=3.5, completedCredits=60, targetCgpa=3.8, remainingCredits=40, scale=4.0.', async () => {
      response = await request.post(`${API_URL}/calculate-gpa`, {
        data: {
          mode: 'target_gpa',
          currentCgpa: 3.5,
          completedCredits: 60,
          targetCgpa: 3.8,
          remainingCredits: 40,
          scale: 4.0,
        },
      });
    });

    await test.step('What we expected: 200 OK.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: requiredGpa = 4.25 (which exceeds the 4.0 max), achievable = false.', async () => {
      const json = await response.json();
      expect(json.requiredGpa).toBe(4.25);
      expect(json.extra.achievable).toBe(false);
    });

    await test.step('Why it got this output: (3.8 × 100 − 3.5 × 60) / 40 = 170 / 40 = 4.25, above the 4.0 scale ceiling.', async () => {});
  });

  test('Persona: Error - target GPA with zero remaining credits', async ({ request }) => {
    await test.step('Why we use this test: A division by zero would be catastrophic; the endpoint must reject remainingCredits=0 before doing the math.', async () => {});

    let response;

    await test.step('What we use: POST with remainingCredits=0.', async () => {
      response = await request.post(`${API_URL}/calculate-gpa`, {
        data: {
          mode: 'target_gpa',
          currentCgpa: 7.0,
          completedCredits: 60,
          targetCgpa: 8.0,
          remainingCredits: 0,
        },
      });
    });

    await test.step('What we expected: 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The error mentions remainingCredits must be greater than zero.', async () => {
      const txt = await response.text();
      expect(txt).toContain('remainingCredits must be greater than zero');
    });

    await test.step('Why it got this output: The validator guards the divisor before the planner runs.', async () => {});
  });

  // ================================================================
  // GROUP 7: cross-cutting validation & shape
  // ================================================================

  test('Persona: Error - missing mode field', async ({ request }) => {
    await test.step('Why we use this test: An empty body must be rejected with a clear message.', async () => {});

    let response;

    await test.step('What we use: POST with an empty JSON body.', async () => {
      response = await request.post(`${API_URL}/calculate-gpa`, { data: {} });
    });

    await test.step('What we expected: 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The error mentions mode is required.', async () => {
      const txt = await response.text();
      expect(txt).toContain('mode is required');
    });

    await test.step('Why it got this output: The dispatcher validates the mode before the switch.', async () => {});
  });

  test('Persona: Error - unsupported mode', async ({ request }) => {
    await test.step('Why we use this test: A typo or hallucinated mode must be rejected with a clear message.', async () => {});

    let response;

    await test.step('What we use: POST with mode="magic".', async () => {
      response = await request.post(`${API_URL}/calculate-gpa`, {
        data: { mode: 'magic' },
      });
    });

    await test.step('What we expected: 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The error mentions unsupported mode.', async () => {
      const txt = await response.text();
      expect(txt).toContain('unsupported mode');
    });

    await test.step('Why it got this output: The dispatcher switch fell through to the default case.', async () => {});
  });

  test('Persona: Error - malformed JSON body', async ({ request }) => {
    await test.step('Why we use this test: A syntactically invalid payload must be caught before any processing.', async () => {});

    let response;

    await test.step('What we use: POST with a raw non-JSON string.', async () => {
      response = await request.post(`${API_URL}/calculate-gpa`, {
        headers: { 'Content-Type': 'application/json' },
        data: 'this is not valid json',
      });
    });

    await test.step('What we expected: 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The error mentions Invalid JSON.', async () => {
      const txt = await response.text();
      expect(txt).toContain('Invalid JSON');
    });

    await test.step('Why it got this output: The json.Decoder failed and the handler returned a 400 immediately.', async () => {});
  });

  test('Persona: Error - wrong HTTP method (GET)', async ({ request }) => {
    await test.step('Why we use this test: The endpoint only accepts POST — GET must be rejected.', async () => {});

    let response;

    await test.step('What we use: GET /calculate-gpa.', async () => {
      response = await request.get(`${API_URL}/calculate-gpa`);
    });

    await test.step('What we expected: 405 Method Not Allowed.', async () => {
      expect(response.status()).toBe(405);
    });

    await test.step('What we get: The error mentions only POST is allowed.', async () => {
      const txt = await response.text();
      expect(txt).toContain('Only POST method is allowed');
    });

    await test.step('Why it got this output: The handler guards the method at the top.', async () => {});
  });

  test('Persona: Response shape — every mode carries formatted and steps', async ({ request }) => {
    await test.step('Why we use this test: The frontend renders `formatted` as the headline value and `steps` in an accordion. Both must be present on every successful response.', async () => {});

    const requests = [
      {
        mode: 'semester_gpa',
        courses: [{ credits: 3, gradePoint: 9 }],
      },
      {
        mode: 'cumulative_cgpa',
        semesters: [{ credits: 20, gpa: 8 }],
      },
      { mode: 'cgpa_to_percentage', cgpa: 8 },
      { mode: 'percentage_to_cgpa', percentage: 76 },
      { mode: 'grade_to_point', grade: 'A' },
      {
        mode: 'target_gpa',
        currentCgpa: 7,
        completedCredits: 60,
        targetCgpa: 8,
        remainingCredits: 40,
      },
    ];

    for (const payload of requests) {
      await test.step(`Verify shape for mode=${payload.mode}`, async () => {
        const res = await request.post(`${API_URL}/calculate-gpa`, { data: payload });
        expect(res.status(), `mode=${payload.mode}`).toBe(200);

        const json = await res.json();
        expect(json.mode, `mode=${payload.mode}`).toBe(payload.mode);
        expect(typeof json.formatted).toBe('string');
        expect(json.formatted.length).toBeGreaterThan(0);
        expect(Array.isArray(json.steps)).toBe(true);
        expect(json.steps.length).toBeGreaterThan(0);
      });
    }

    await test.step('Why it got this output: Every converter branch populates Formatted and Steps by construction.', async () => {});
  });

});