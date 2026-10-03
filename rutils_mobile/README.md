# Rutils Mobile Application

A cross-platform mobile application built using **Flutter & Dart**, designed for both **Android** and **iOS** with responsive local **Web Preview** capabilities.

The app uses the same backend architecture as the web platform, configured for **Local development (`http://localhost:8080`)** and **Production (`https://utils.api.srilakshmiretail.in`)**.

---

## 📱 Features & Capabilities

### 1. 🔐 Authentication & Session Security (Strict Login First)
- **First Screen Experience**: Premium Blue-themed branding with animated entrance, logo glow, and input validation.
- **Full Auth Lifecycle**:
  - **Login**: Email & Password authentication returning JWT Bearer token.
  - **Registration**: User onboarding with live password complexity indicators (min 8 chars, mixed case, numbers, special characters).
  - **Forgot Password**: Password reset request with Google SMTP / development console logging fallback.
  - **Forced & Voluntary Password Reset**: Modal prompt when temporary password is used or changed voluntarily.
  - **Profile & Health Check**: User profile details, live backend connectivity latency ping test, and secure sign-out.

### 2. 🔄 Conversion Hub (`ConversionHubScreen`)
- **File Converter**: Universal file conversion supporting `CSV`, `JSON`, `TXT`, `DOCX`, `PDF`, `JPG`, and `PNG` with automatic file extension detection and direct Share/Save actions.
- **PDF Size Optimizer**: Compress or expand PDF documents to target file size (KB) or percentage (%).
- **Unit Converter**: Physical unit conversions across 8 categories:
  - Length (Meters, Kilometers, Centimeters, Millimeters, Miles, Yards, Feet, Inches)
  - Weight (Kilograms, Grams, Milligrams, Pounds, Ounces)
  - Volume (Liters, Milliliters, Gallons, Quarts, Pints, Fluid Ounces)
  - Area (Square Meters, Square Kilometers, Hectares, Acres, Square Feet, Square Miles)
  - Time (Seconds, Minutes, Hours, Days, Weeks)
  - Temperature (Celsius, Fahrenheit, Kelvin)
  - Speed (m/s, km/h, mph, ft/s, knots)
  - Data (Bytes, KB, MB, GB, TB, PB, bits)
- **Time Zone Converter**: Multi-zone converter with Date/Time picker, automatic DST offset calculations, UTC mapping, and Next/Prev day indicators.
- **Railway Time Converter**: Dual-way 12-hour AM/PM ⇄ 24-hour military/railway time converter.
- **Numeral System Converter**: Decimal, Binary, Octal, and Hexadecimal converter with copy-to-clipboard.
- **OCR Text Extraction**: Optical character recognition and text parsing for PDFs and images.

### 3. 🧮 Calculations Hub (`CalculationsHubScreen`)
- **BMI Calculator**: Interactive Body Mass Index computation with color-coded health ranges (Underweight, Normal, Overweight, Obese).
- **Age Calculator**: Precise chronological age calculation in years, months, and days; next birthday countdown; and total lifespan statistics (hours, days, weeks).
- **Percentage Calculator**: 6 operation modes including percentage of number, what percent is X of Y, percentage increase/decrease, and step-by-step breakdown.
- **Loan EMI Calculator**: Monthly EMI calculation, total interest, principal vs. interest breakdown with interactive animated **Pie Chart (`fl_chart`)**, and expandable monthly repayment amortization schedule.
- **Tax / GST Calculator**: GST, VAT, and Income tax calculation with Add Tax, Remove Tax, Split GST (CGST/SGST), Reverse GST, and Old vs. New tax regime support.

### 4. 📈 Markets Hub (`Nifty50Screen`)
- **Live NIFTY 50 Constituent Browser**: Real-time constituents list with company name, ticker symbol, industry tags, series (`EQ`), and ISIN numbers.
- **Search & Filter**: Instant search by symbol, company, or ISIN code; horizontal industry category filtering; pull-to-refresh.

---

## 🎨 UI & Design System

- **Electric & Royal Blue Theme**: `#1E40AF` (Royal Blue), `#2563EB` (Electric Blue), `#0EA5E9` (Sky Cyan), `#0F172A` (Midnight Navy).
- **Glassmorphism & Frosted Surfaces**: Translucent cards with `BackdropFilter` gaussian blur (`sigma: 10`) and subtle hairline borders.
- **Tactile Micro-Interactions**: Bouncy scale-down button feedback (`0.96x`), smooth transitions via `flutter_animate`.
- **Responsive Adaptive Layout**: Optimized for mobile screens, tablets, and web viewports.

---

## 🛠️ Management & Build Scripts

The repository includes executable scripts located in the root folder and under `rutils_mobile/scripts/`:

| Script | Purpose | Command |
| :--- | :--- | :--- |
| **Interactive Manager** | Comprehensive terminal dashboard with colored menu | `./manage_mobile.sh` |
| **Start Web Preview** | Runs the local responsive web preview on port `8085` | `./start_mobile_web.sh [port] [local\|prod] [daemon]` |
| **Stop Web Preview** | Gracefully terminates any running web preview instance | `./stop_mobile_web.sh` |
| **Build Android** | Compiles Debug APK, Release APK, or Play Store AAB | `./build_mobile_android.sh [all\|apk\|aab\|debug]` |
| **Build iOS** | Compiles iOS Release Runner and bundle assets | `./rutils_mobile/scripts/build_ios.sh` |

---

## 🚀 Quick Start Guide

### 1. View Local Mobile Web Preview
Run the local preview runner to test the mobile application directly in your browser:
```bash
./start_mobile_web.sh 8085 local daemon
```
Open **`http://localhost:8085`** in your browser.

To stop the preview:
```bash
./stop_mobile_web.sh
```

### 2. Run Interactive Dashboard
```bash
./manage_mobile.sh
```

### 3. Build Production Artifacts

#### A. Google Play Store Deployable App Bundle (AAB):
```bash
./build_mobile_android.sh aab
```
*Output: `rutils_mobile/build/app/outputs/bundle/release/app-release.aab`*

#### B. Direct Sideload / Distro Release APK:
```bash
./build_mobile_android.sh apk
```
*Output: `rutils_mobile/build/app/outputs/flutter-apk/app-release.apk`*

#### C. Run on Physical Device or Emulator:
```bash
cd rutils_mobile
flutter run
```
