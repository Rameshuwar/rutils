// ============================================================
// AUTH-UI.TS — Auth modal render + event handling
// ============================================================

import {
  getAuthState,
  saveAuth,
  clearAuth,
  isLoggedIn,
  apiRegister,
  apiLogin,
  apiLogout,
  apiForgotPassword,
  apiResetPassword,
  apiChangePassword,
} from './auth';

// ─── Types ──────────────────────────────────────────────────
type AuthScreen = 'login' | 'register' | 'forgot' | 'reset' | 'change';

// ─── Inject modal HTML into the page ─────────────────────────
function injectModal(): void {
  const tpl = document.createElement('div');
  tpl.innerHTML = `
<!-- ╔══════════════════════════════════════════════════════╗ -->
<!-- ║  AUTH MODAL OVERLAY                                  ║ -->
<!-- ╚══════════════════════════════════════════════════════╝ -->
<div id="auth-overlay"
     class="fixed inset-0 z-[200] hidden items-center justify-center bg-black/60 backdrop-blur-sm p-4"
     role="dialog" aria-modal="true" aria-labelledby="auth-modal-title">

  <!-- Card -->
  <div class="relative w-full max-w-md bg-white rounded-2xl shadow-2xl overflow-hidden">

    <!-- Gradient header bar -->
    <div id="auth-modal-header" class="bg-gradient-to-r from-indigo-600 to-violet-600 px-8 py-7 text-white">
      <div class="flex items-center justify-between">
        <div>
          <h2 id="auth-modal-title" class="text-2xl font-bold tracking-tight"></h2>
          <p id="auth-modal-subtitle" class="text-indigo-200 text-sm mt-1"></p>
        </div>
        <button id="auth-close-btn"
          class="p-2 rounded-xl hover:bg-white/20 transition-colors"
          aria-label="Close">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
          </svg>
        </button>
      </div>
    </div>

    <!-- Body -->
    <div class="px-8 py-7 space-y-5">

      <!-- Global error / success banner -->
      <div id="auth-banner" class="hidden rounded-xl px-4 py-3 text-sm font-medium flex items-start gap-2"></div>

      <!-- ── SCREEN: LOGIN ──────────────────────────────── -->
      <form id="auth-form-login" class="auth-form space-y-4" novalidate>
        <div>
          <label class="block text-sm font-semibold text-gray-700 mb-1.5">Email address</label>
          <input type="email" id="login-email" autocomplete="email"
            placeholder="jane@example.com"
            class="auth-input w-full" required />
        </div>
        <div>
          <label class="block text-sm font-semibold text-gray-700 mb-1.5">Password</label>
          <div class="relative">
            <input type="password" id="login-password" autocomplete="current-password"
              placeholder="••••••••"
              class="auth-input w-full pr-10" required />
            <button type="button" class="pw-toggle absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600" tabindex="-1">
              <svg class="w-4 h-4 eye-show" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"/></svg>
              <svg class="w-4 h-4 eye-hide hidden" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21"/></svg>
            </button>
          </div>
          <button type="button" id="goto-forgot"
            class="mt-1.5 text-xs text-indigo-600 hover:text-indigo-800 hover:underline font-medium">
            Forgot your password?
          </button>
        </div>
        <button type="submit" class="auth-submit-btn w-full">
          <span class="btn-label">Sign In</span>
          <span class="btn-spinner hidden">
            <svg class="animate-spin h-4 w-4 inline" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/></svg>
            Signing in…
          </span>
        </button>
        <p class="text-center text-sm text-gray-500">
          Don't have an account?
          <button type="button" id="goto-register"
            class="text-indigo-600 hover:text-indigo-800 font-semibold hover:underline">Sign up</button>
        </p>
      </form>

      <!-- ── SCREEN: REGISTER ───────────────────────────── -->
      <form id="auth-form-register" class="auth-form hidden space-y-4" novalidate>
        <div>
          <label class="block text-sm font-semibold text-gray-700 mb-1.5">Full name</label>
          <input type="text" id="reg-name" autocomplete="name"
            placeholder="Jane Doe"
            class="auth-input w-full" required />
        </div>
        <div>
          <label class="block text-sm font-semibold text-gray-700 mb-1.5">Email address</label>
          <input type="email" id="reg-email" autocomplete="email"
            placeholder="jane@example.com"
            class="auth-input w-full" required />
        </div>
        <div>
          <label class="block text-sm font-semibold text-gray-700 mb-1.5">Password</label>
          <div class="relative">
            <input type="password" id="reg-password" autocomplete="new-password"
              placeholder="Min 8 chars, upper, lower, number, special"
              class="auth-input w-full pr-10" required />
            <button type="button" class="pw-toggle absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600" tabindex="-1">
              <svg class="w-4 h-4 eye-show" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"/></svg>
              <svg class="w-4 h-4 eye-hide hidden" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21"/></svg>
            </button>
          </div>
          <!-- Password strength bar -->
          <div class="mt-2 space-y-1.5">
            <div class="h-1.5 bg-gray-200 rounded-full overflow-hidden">
              <div id="pw-strength-bar" class="h-full rounded-full transition-all duration-300" style="width:0%"></div>
            </div>
            <p id="pw-strength-label" class="text-xs text-gray-400"></p>
            <ul class="grid grid-cols-2 gap-x-4 gap-y-0.5 text-xs">
              <li id="req-len"    class="req-item flex items-center gap-1 text-gray-400"><span class="req-dot">○</span> 8+ characters</li>
              <li id="req-upper"  class="req-item flex items-center gap-1 text-gray-400"><span class="req-dot">○</span> Uppercase letter</li>
              <li id="req-lower"  class="req-item flex items-center gap-1 text-gray-400"><span class="req-dot">○</span> Lowercase letter</li>
              <li id="req-digit"  class="req-item flex items-center gap-1 text-gray-400"><span class="req-dot">○</span> Number</li>
              <li id="req-special" class="req-item flex items-center gap-1 text-gray-400"><span class="req-dot">○</span> Special character</li>
            </ul>
          </div>
        </div>
        <div>
          <label class="block text-sm font-semibold text-gray-700 mb-1.5">Confirm password</label>
          <div class="relative">
            <input type="password" id="reg-confirm" autocomplete="new-password"
              placeholder="Re-enter password"
              class="auth-input w-full pr-10" required />
            <button type="button" class="pw-toggle absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600" tabindex="-1">
              <svg class="w-4 h-4 eye-show" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"/></svg>
              <svg class="w-4 h-4 eye-hide hidden" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21"/></svg>
            </button>
          </div>
        </div>
        <button type="submit" class="auth-submit-btn w-full">
          <span class="btn-label">Create Account</span>
          <span class="btn-spinner hidden">
            <svg class="animate-spin h-4 w-4 inline" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/></svg>
            Creating account…
          </span>
        </button>
        <p class="text-center text-sm text-gray-500">
          Already have an account?
          <button type="button" id="goto-login-from-register"
            class="text-indigo-600 hover:text-indigo-800 font-semibold hover:underline">Sign in</button>
        </p>
      </form>

      <!-- ── SCREEN: FORGOT PASSWORD ────────────────────── -->
      <form id="auth-form-forgot" class="auth-form hidden space-y-4" novalidate>
        <p class="text-sm text-gray-600 leading-relaxed">
          Enter the email linked to your account and we'll send you a temporary password to log in with.
        </p>
        <div>
          <label class="block text-sm font-semibold text-gray-700 mb-1.5">Email address</label>
          <input type="email" id="forgot-email" autocomplete="email"
            placeholder="jane@example.com"
            class="auth-input w-full" required />
        </div>
        <button type="submit" class="auth-submit-btn w-full">
          <span class="btn-label">Send Temporary Password</span>
          <span class="btn-spinner hidden">
            <svg class="animate-spin h-4 w-4 inline" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/></svg>
            Sending…
          </span>
        </button>
        <p class="text-center text-sm text-gray-500">
          <button type="button" id="goto-login-from-forgot"
            class="text-indigo-600 hover:text-indigo-800 font-semibold hover:underline">← Back to sign in</button>
        </p>
      </form>

      <!-- ── SCREEN: RESET PASSWORD (forced after temp pw) ─ -->
      <form id="auth-form-reset" class="auth-form hidden space-y-4" novalidate>
        <div class="bg-amber-50 border border-amber-200 rounded-xl px-4 py-3 flex items-start gap-2">
          <svg class="w-4 h-4 text-amber-500 mt-0.5 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z"/></svg>
          <p class="text-sm text-amber-700 font-medium">You logged in with a temporary password. Please set a permanent password to continue.</p>
        </div>
        <div>
          <label class="block text-sm font-semibold text-gray-700 mb-1.5">New password</label>
          <div class="relative">
            <input type="password" id="reset-new-password" autocomplete="new-password"
              placeholder="Min 8 chars, upper, lower, number, special"
              class="auth-input w-full pr-10" required />
            <button type="button" class="pw-toggle absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600" tabindex="-1">
              <svg class="w-4 h-4 eye-show" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"/></svg>
              <svg class="w-4 h-4 eye-hide hidden" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21"/></svg>
            </button>
          </div>
        </div>
        <div>
          <label class="block text-sm font-semibold text-gray-700 mb-1.5">Confirm new password</label>
          <div class="relative">
            <input type="password" id="reset-confirm-password" autocomplete="new-password"
              placeholder="Re-enter new password"
              class="auth-input w-full pr-10" required />
            <button type="button" class="pw-toggle absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600" tabindex="-1">
              <svg class="w-4 h-4 eye-show" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"/></svg>
              <svg class="w-4 h-4 eye-hide hidden" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21"/></svg>
            </button>
          </div>
        </div>
        <button type="submit" class="auth-submit-btn w-full">
          <span class="btn-label">Set New Password</span>
          <span class="btn-spinner hidden">
            <svg class="animate-spin h-4 w-4 inline" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/></svg>
            Saving…
          </span>
        </button>
      </form>

      <!-- ── SCREEN: CHANGE PASSWORD (logged in) ─────────── -->
      <form id="auth-form-change" class="auth-form hidden space-y-4" novalidate>
        <div>
          <label class="block text-sm font-semibold text-gray-700 mb-1.5">Current password</label>
          <div class="relative">
            <input type="password" id="change-current" autocomplete="current-password"
              placeholder="Your current password"
              class="auth-input w-full pr-10" required />
            <button type="button" class="pw-toggle absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600" tabindex="-1">
              <svg class="w-4 h-4 eye-show" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"/></svg>
              <svg class="w-4 h-4 eye-hide hidden" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21"/></svg>
            </button>
          </div>
        </div>
        <div>
          <label class="block text-sm font-semibold text-gray-700 mb-1.5">New password</label>
          <div class="relative">
            <input type="password" id="change-new" autocomplete="new-password"
              placeholder="Min 8 chars, upper, lower, number, special"
              class="auth-input w-full pr-10" required />
            <button type="button" class="pw-toggle absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600" tabindex="-1">
              <svg class="w-4 h-4 eye-show" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"/></svg>
              <svg class="w-4 h-4 eye-hide hidden" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21"/></svg>
            </button>
          </div>
        </div>
        <div>
          <label class="block text-sm font-semibold text-gray-700 mb-1.5">Confirm new password</label>
          <div class="relative">
            <input type="password" id="change-confirm" autocomplete="new-password"
              placeholder="Re-enter new password"
              class="auth-input w-full pr-10" required />
            <button type="button" class="pw-toggle absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600" tabindex="-1">
              <svg class="w-4 h-4 eye-show" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"/></svg>
              <svg class="w-4 h-4 eye-hide hidden" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21"/></svg>
            </button>
          </div>
        </div>
        <button type="submit" class="auth-submit-btn w-full">
          <span class="btn-label">Change Password</span>
          <span class="btn-spinner hidden">
            <svg class="animate-spin h-4 w-4 inline" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/></svg>
            Changing…
          </span>
        </button>
      </form>

    </div><!-- /body -->
  </div><!-- /card -->
</div><!-- /overlay -->
`;
  document.body.appendChild(tpl.children[0]);
}

// ─── Screen config ─────────────────────────────────────────
interface ScreenConfig {
  title: string;
  subtitle: string;
}

const SCREEN_CONFIG: Record<AuthScreen, ScreenConfig> = {
  login:    { title: 'Welcome back',        subtitle: 'Sign in to your Rutils account' },
  register: { title: 'Create account',      subtitle: 'Join Rutils — it\'s free' },
  forgot:   { title: 'Forgot password?',    subtitle: 'We\'ll email you a temporary password' },
  reset:    { title: 'Set new password',    subtitle: 'Choose a strong permanent password' },
  change:   { title: 'Change password',     subtitle: 'Update your account password' },
};

// ─── Helpers ───────────────────────────────────────────────
function showBanner(message: string, isError: boolean): void {
  const el = document.getElementById('auth-banner')!;
  el.className = [
    'rounded-xl px-4 py-3 text-sm font-medium flex items-start gap-2',
    isError ? 'bg-red-50 text-red-700 border border-red-200' : 'bg-green-50 text-green-700 border border-green-200',
  ].join(' ');
  el.innerHTML = isError
    ? `<svg class="w-4 h-4 mt-0.5 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/></svg><span>${message}</span>`
    : `<svg class="w-4 h-4 mt-0.5 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/></svg><span>${message}</span>`;
  el.classList.remove('hidden');
}

function hideBanner(): void {
  document.getElementById('auth-banner')!.classList.add('hidden');
}

function setLoading(form: HTMLFormElement, loading: boolean): void {
  const btn = form.querySelector<HTMLButtonElement>('button[type="submit"]')!;
  const label = btn.querySelector('.btn-label')!;
  const spinner = btn.querySelector('.btn-spinner')!;
  btn.disabled = loading;
  label.classList.toggle('hidden', loading);
  spinner.classList.toggle('hidden', !loading);
}

function showScreen(screen: AuthScreen): void {
  hideBanner();

  // Update header
  const cfg = SCREEN_CONFIG[screen];
  document.getElementById('auth-modal-title')!.textContent = cfg.title;
  document.getElementById('auth-modal-subtitle')!.textContent = cfg.subtitle;

  // Show/hide forms
  document.querySelectorAll('.auth-form').forEach(f => f.classList.add('hidden'));
  document.getElementById(`auth-form-${screen}`)!.classList.remove('hidden');
}

// ─── Password strength meter ──────────────────────────────
function measureStrength(pw: string): { score: number; label: string; color: string; width: string } {
  let score = 0;
  if (pw.length >= 8)                   score++;
  if (/[A-Z]/.test(pw))                 score++;
  if (/[a-z]/.test(pw))                 score++;
  if (/\d/.test(pw))                    score++;
  if (/[^A-Za-z0-9]/.test(pw))         score++;

  const map: [number, string, string, string][] = [
    [0, '',        '',              '0%'],
    [1, 'Weak',    'bg-red-400',    '20%'],
    [2, 'Fair',    'bg-orange-400', '40%'],
    [3, 'Good',    'bg-yellow-400', '60%'],
    [4, 'Strong',  'bg-blue-500',   '80%'],
    [5, 'Excellent', 'bg-green-500','100%'],
  ];
  const [, label, color, width] = map[score];
  return { score, label, color, width };
}

function updateStrengthUI(pw: string): void {
  const bar   = document.getElementById('pw-strength-bar')!;
  const lbl   = document.getElementById('pw-strength-label')!;
  const { label, color, width } = measureStrength(pw);

  bar.className = `h-full rounded-full transition-all duration-300 ${color}`;
  bar.style.width = width;
  lbl.textContent = label ? `Strength: ${label}` : '';

  const checks: [string, boolean][] = [
    ['req-len',     pw.length >= 8],
    ['req-upper',   /[A-Z]/.test(pw)],
    ['req-lower',   /[a-z]/.test(pw)],
    ['req-digit',   /\d/.test(pw)],
    ['req-special', /[^A-Za-z0-9]/.test(pw)],
  ];
  checks.forEach(([id, ok]) => {
    const li = document.getElementById(id)!;
    const dot = li.querySelector('.req-dot')!;
    li.className = `req-item flex items-center gap-1 ${ok ? 'text-green-600 font-medium' : 'text-gray-400'}`;
    dot.textContent = ok ? '✓' : '○';
  });
}

// ─── Password visibility toggles ─────────────────────────
function initPasswordToggles(): void {
  document.querySelectorAll('.pw-toggle').forEach(btn => {
    btn.addEventListener('click', () => {
      const wrapper = btn.closest('.relative')!;
      const input = wrapper.querySelector<HTMLInputElement>('input')!;
      const show = wrapper.querySelector('.eye-show')!;
      const hide = wrapper.querySelector('.eye-hide')!;
      const isText = input.type === 'text';
      input.type = isText ? 'password' : 'text';
      show.classList.toggle('hidden', !isText);
      hide.classList.toggle('hidden', isText);
    });
  });
}

// ─── Modal open/close ─────────────────────────────────────
let currentScreen: AuthScreen = 'login';

function openModal(screen: AuthScreen = 'login'): void {
  currentScreen = screen;
  const overlay = document.getElementById('auth-overlay')!;
  overlay.classList.remove('hidden');
  overlay.classList.add('flex');
  showScreen(screen);
  document.body.style.overflow = 'hidden';

  // Focus first input
  setTimeout(() => {
    const input = document.querySelector<HTMLInputElement>(`#auth-form-${screen} input`);
    input?.focus();
  }, 50);
}

function closeModal(): void {
  const overlay = document.getElementById('auth-overlay')!;
  overlay.classList.add('hidden');
  overlay.classList.remove('flex');
  document.body.style.overflow = '';
  hideBanner();

  // Reset all forms
  document.querySelectorAll<HTMLFormElement>('.auth-form').forEach(f => f.reset());
  // Reset strength bar
  const bar = document.getElementById('pw-strength-bar');
  if (bar) { bar.style.width = '0%'; bar.className = 'h-full rounded-full transition-all duration-300'; }
  const lbl = document.getElementById('pw-strength-label');
  if (lbl) lbl.textContent = '';
}

// ─── Form submissions ─────────────────────────────────────
async function handleLogin(e: Event): Promise<void> {
  e.preventDefault();
  const form = document.getElementById('auth-form-login') as HTMLFormElement;
  const email    = (document.getElementById('login-email') as HTMLInputElement).value.trim();
  const password = (document.getElementById('login-password') as HTMLInputElement).value;

  hideBanner();
  setLoading(form, true);

  try {
    const res = await apiLogin(email, password);
    saveAuth(res.token, res.user);
    updateNavButton();

    if (res.must_reset_password) {
      // Store token for reset use, go to reset screen
      showScreen('reset');
      showBanner('Temporary password accepted. Please set a new permanent password.', false);
    } else {
      closeModal();
      showToast(`Welcome back, ${res.user.name}! 👋`);
    }
  } catch (err) {
    showBanner((err as Error).message, true);
  } finally {
    setLoading(form, false);
  }
}

async function handleRegister(e: Event): Promise<void> {
  e.preventDefault();
  const form = document.getElementById('auth-form-register') as HTMLFormElement;
  const name     = (document.getElementById('reg-name') as HTMLInputElement).value.trim();
  const email    = (document.getElementById('reg-email') as HTMLInputElement).value.trim();
  const password = (document.getElementById('reg-password') as HTMLInputElement).value;
  const confirm  = (document.getElementById('reg-confirm') as HTMLInputElement).value;

  hideBanner();
  setLoading(form, true);

  try {
    await apiRegister(name, email, password, confirm);
    // Auto-login after register
    const res = await apiLogin(email, password);
    saveAuth(res.token, res.user);
    updateNavButton();
    closeModal();
    showToast(`Account created! Welcome to Rutils, ${res.user.name} 🎉`);
  } catch (err) {
    showBanner((err as Error).message, true);
  } finally {
    setLoading(form, false);
  }
}

async function handleForgotPassword(e: Event): Promise<void> {
  e.preventDefault();
  const form = document.getElementById('auth-form-forgot') as HTMLFormElement;
  const email = (document.getElementById('forgot-email') as HTMLInputElement).value.trim();

  hideBanner();
  setLoading(form, true);

  try {
    await apiForgotPassword(email);
    showBanner('A temporary password has been sent to your email. Use it to log in.', false);
    setTimeout(() => showScreen('login'), 3000);
  } catch (err) {
    showBanner((err as Error).message, true);
  } finally {
    setLoading(form, false);
  }
}

async function handleResetPassword(e: Event): Promise<void> {
  e.preventDefault();
  const form      = document.getElementById('auth-form-reset') as HTMLFormElement;
  const newPw     = (document.getElementById('reset-new-password') as HTMLInputElement).value;
  const confirmPw = (document.getElementById('reset-confirm-password') as HTMLInputElement).value;

  hideBanner();
  setLoading(form, true);

  try {
    const { token } = getAuthState();
    if (!token) throw new Error('No active session. Please log in again.');
    await apiResetPassword(token, newPw, confirmPw);
    showBanner('Password updated successfully!', false);
    setTimeout(() => {
      closeModal();
      showToast('Your password has been set. You\'re all set! ✅');
    }, 1500);
  } catch (err) {
    showBanner((err as Error).message, true);
  } finally {
    setLoading(form, false);
  }
}

async function handleChangePassword(e: Event): Promise<void> {
  e.preventDefault();
  const form      = document.getElementById('auth-form-change') as HTMLFormElement;
  const current   = (document.getElementById('change-current') as HTMLInputElement).value;
  const newPw     = (document.getElementById('change-new') as HTMLInputElement).value;
  const confirmPw = (document.getElementById('change-confirm') as HTMLInputElement).value;

  hideBanner();
  setLoading(form, true);

  try {
    const { token } = getAuthState();
    if (!token) throw new Error('You must be logged in to change your password.');
    await apiChangePassword(token, current, newPw, confirmPw);
    showBanner('Password changed! A confirmation email has been sent.', false);
    setTimeout(() => {
      closeModal();
      showToast('Password changed successfully ✅');
    }, 1800);
  } catch (err) {
    showBanner((err as Error).message, true);
  } finally {
    setLoading(form, false);
  }
}

// ─── Sidebar / Header login button ───────────────────────
function updateNavButton(): void {
  const { user } = getAuthState();
  const loggedIn = isLoggedIn();

  // Update all [data-auth-btn] elements
  document.querySelectorAll<HTMLElement>('[data-auth-btn]').forEach(el => {
    if (loggedIn && user) {
      // Show user avatar / name button
      el.innerHTML = `
        <div class="flex items-center gap-2 w-full text-left">
          <div class="w-8 h-8 rounded-full bg-indigo-400 flex items-center justify-center text-white font-bold text-xs shrink-0">
            ${user.name.charAt(0).toUpperCase()}
          </div>
          <div class="min-w-0">
            <p class="text-sm font-semibold text-white leading-tight truncate">${user.name}</p>
            <p class="text-xs text-indigo-300 leading-tight truncate">${user.email}</p>
          </div>
        </div>
      `;
      el.setAttribute('data-logged-in', 'true');
    } else {
      el.innerHTML = `
        <svg class="w-4 h-4 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
            d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"/>
        </svg>
        <span class="font-semibold">Sign In</span>
      `;
      el.setAttribute('data-logged-in', 'false');
    }
  });

  // Update user action menu visibility
  document.querySelectorAll<HTMLElement>('[data-auth-menu]').forEach(el => {
    el.classList.toggle('hidden', !loggedIn);
  });
}

// ─── User action menu ─────────────────────────────────────
let menuOpen = false;

function createUserMenu(): void {
  const existing = document.getElementById('auth-user-menu');
  if (existing) existing.remove();

  const menu = document.createElement('div');
  menu.id = 'auth-user-menu';
  menu.className = 'hidden absolute bottom-20 left-4 right-4 bg-white rounded-xl shadow-2xl border border-gray-100 overflow-hidden z-[100] md:left-auto md:right-auto md:w-56';
  menu.innerHTML = `
    <div class="px-4 py-3 border-b border-gray-100 bg-gray-50">
      <p class="text-xs text-gray-500 uppercase font-semibold tracking-wide">Account</p>
    </div>
    <button id="menu-change-pw"
      class="w-full flex items-center gap-3 px-4 py-3 text-sm text-gray-700 hover:bg-indigo-50 hover:text-indigo-700 transition-colors text-left">
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
          d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z"/>
      </svg>
      Change Password
    </button>
    <div class="border-t border-gray-100"></div>
    <button id="menu-logout"
      class="w-full flex items-center gap-3 px-4 py-3 text-sm text-red-600 hover:bg-red-50 transition-colors text-left">
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
          d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"/>
      </svg>
      Sign Out
    </button>
  `;

  document.getElementById('sidebar')!.appendChild(menu);

  document.getElementById('menu-change-pw')!.addEventListener('click', () => {
    closeUserMenu();
    openModal('change');
  });

  document.getElementById('menu-logout')!.addEventListener('click', async () => {
    closeUserMenu();
    try {
      const { token } = getAuthState();
      if (token) await apiLogout(token);
    } catch { /* ignore logout errors */ }
    clearAuth();
    updateNavButton();
    showToast('Signed out successfully 👋');
  });
}

function openUserMenu(): void {
  const menu = document.getElementById('auth-user-menu');
  if (menu) {
    menuOpen = !menuOpen;
    menu.classList.toggle('hidden', !menuOpen);
  }
}

function closeUserMenu(): void {
  menuOpen = false;
  document.getElementById('auth-user-menu')?.classList.add('hidden');
}

// ─── Toast notifications ──────────────────────────────────
function showToast(message: string): void {
  const existing = document.getElementById('auth-toast');
  if (existing) existing.remove();

  const toast = document.createElement('div');
  toast.id = 'auth-toast';
  toast.className = [
    'fixed bottom-6 left-1/2 -translate-x-1/2 z-[300]',
    'bg-gray-900 text-white text-sm font-medium px-5 py-3 rounded-full shadow-2xl',
    'flex items-center gap-2',
    'opacity-0 translate-y-4 transition-all duration-300',
  ].join(' ');
  toast.innerHTML = `<svg class="w-4 h-4 text-green-400 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/></svg><span>${message}</span>`;
  document.body.appendChild(toast);

  // Animate in
  requestAnimationFrame(() => {
    requestAnimationFrame(() => {
      toast.classList.remove('opacity-0', 'translate-y-4');
    });
  });

  setTimeout(() => {
    toast.classList.add('opacity-0', 'translate-y-4');
    setTimeout(() => toast.remove(), 300);
  }, 3500);
}

// ─── Init ─────────────────────────────────────────────────
export function initAuth(): void {
  injectModal();

  // ─ Overlay close on backdrop click ─
  const overlay = document.getElementById('auth-overlay')!;
  overlay.addEventListener('click', (e) => {
    if (e.target === overlay) closeModal();
  });
  document.getElementById('auth-close-btn')!.addEventListener('click', closeModal);

  // ─ Keyboard close ─
  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') { closeModal(); closeUserMenu(); }
  });

  // ─ Screen navigation links ─
  document.getElementById('goto-register')!.addEventListener('click', () => openModal('register'));
  document.getElementById('goto-forgot')!.addEventListener('click', () => openModal('forgot'));
  document.getElementById('goto-login-from-register')!.addEventListener('click', () => openModal('login'));
  document.getElementById('goto-login-from-forgot')!.addEventListener('click', () => openModal('login'));

  // ─ Form submissions ─
  document.getElementById('auth-form-login')!.addEventListener('submit', handleLogin);
  document.getElementById('auth-form-register')!.addEventListener('submit', handleRegister);
  document.getElementById('auth-form-forgot')!.addEventListener('submit', handleForgotPassword);
  document.getElementById('auth-form-reset')!.addEventListener('submit', handleResetPassword);
  document.getElementById('auth-form-change')!.addEventListener('submit', handleChangePassword);

  // ─ Password strength on register ─
  document.getElementById('reg-password')!.addEventListener('input', (e) => {
    updateStrengthUI((e.target as HTMLInputElement).value);
  });

  // ─ Password visibility toggles ─
  initPasswordToggles();

  // ─ Sidebar sign-in button ─
  document.querySelectorAll<HTMLElement>('[data-auth-btn]').forEach(btn => {
    btn.addEventListener('click', () => {
      if (btn.getAttribute('data-logged-in') === 'true') {
        openUserMenu();
      } else {
        openModal('login');
      }
    });
  });

  // ─ Create user dropdown menu ─
  createUserMenu();

  // ─ Close menu when clicking outside ─
  document.addEventListener('click', (e) => {
    const menu = document.getElementById('auth-user-menu');
    const btns = document.querySelectorAll('[data-auth-btn]');
    let clickedBtn = false;
    btns.forEach(b => { if (b.contains(e.target as Node)) clickedBtn = true; });
    if (!clickedBtn && menu && !menu.contains(e.target as Node)) {
      closeUserMenu();
    }
  });

  // ─ Initial state ─
  updateNavButton();
}
