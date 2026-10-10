// ============================================================
// GPA-SERVICE.TS — GPA / CGPA calculator API client & types
//
// Mirrors the backend's /calculate-gpa contract exactly.
// One entry point: calculateGPA(req).
//
// This file has no DOM dependencies — it can be imported by any
// UI controller (gpa-ui.ts) or by tests without side effects.
// ============================================================

import { apiPost } from './auth';

// ─── Request types ──────────────────────────────────────────

/** A single course inside a semester_gpa request. */
export interface GPACourse {
  name?: string;
  credits: number;
  gradePoint: number;
  grade?: string;
}

/** A single semester inside a cumulative_cgpa request. */
export interface GPASemester {
  name?: string;
  credits: number;
  gpa: number;
}

/**
 * The unified request body for /calculate-gpa.
 *
 * Only the fields relevant to the chosen `mode` need to be set.
 * The backend ignores fields it does not use.
 */
export interface GPARequest {
  mode:
    | 'semester_gpa'
    | 'cumulative_cgpa'
    | 'cgpa_to_percentage'
    | 'percentage_to_cgpa'
    | 'grade_to_point'
    | 'target_gpa';

  // mode = semester_gpa
  courses?: GPACourse[];

  // mode = cumulative_cgpa
  semesters?: GPASemester[];

  // mode = cgpa_to_percentage | percentage_to_cgpa
  cgpa?: number;
  percentage?: number;
  multiplier?: number;

  // mode = grade_to_point
  grade?: string;
  gradingScale?: string;   // "10" | "4" | "5"

  // mode = target_gpa
  currentCgpa?: number;
  completedCredits?: number;
  targetCgpa?: number;
  remainingCredits?: number;
  scale?: number;
}

// ─── Response types ─────────────────────────────────────────

/** One row in the semester_gpa breakdown. */
export interface GPACourseResult {
  name?: string;
  credits: number;
  gradePoint: number;
  weighted: number;
}

/** One row in the cumulative_cgpa breakdown. */
export interface GPASemesterResult {
  name?: string;
  credits: number;
  gpa: number;
  weighted: number;
}

/** The unified response body from /calculate-gpa. */
export interface GPAResponse {
  mode: string;

  gpa?: number;
  cgpa?: number;
  percentage?: number;
  gradePoint?: number;
  requiredGpa?: number;

  currentCgpa?: number;
  targetCgpa?: number;
  totalCredits?: number;
  totalWeighted?: number;
  multiplier?: number;

  formatted: string;

  breakdown?: GPACourseResult[];
  semesters?: GPASemesterResult[];
  steps?: string[];
  extra?: Record<string, unknown>;
}

// ─── API call ───────────────────────────────────────────────

/**
 * POST the request to /calculate-gpa and return the parsed
 * response. Throws an Error with a human-readable message on
 * any non-2xx response.
 *
 * `apiPost` from auth.ts handles:
 *   - resolving the correct base URL (local vs prod)
 *   - attaching the Authorization header when a token exists
 *   - unwrapping the backend's {error: "..."} envelope into a
 *     thrown Error, so callers just do try/catch
 */
export async function calculateGPA(req: GPARequest): Promise<GPAResponse> {
  return apiPost<GPAResponse>('/calculate-gpa', req);
}