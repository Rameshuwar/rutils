import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';

/// Renders the Flutter mobile application in a clean, tall mobile layout
/// centered on desktop screens with 1:1 native rendering for maximum responsiveness.
///
/// Features:
/// - Longer screen view: Utilizes 100% full vertical height of the browser
/// - Native 1:1 pixel rendering: Zero scaling transforms or heavy drop-shadows
/// - Clean mobile camera notch at top with 38px top safe-area padding
/// - Minimal home indicator at bottom
/// - Automatically takes full screen on mobile devices (width <= 500)
class MobileDeviceFrame extends StatelessWidget {
  final Widget child;

  const MobileDeviceFrame({super.key, required this.child});

  @override
  Widget build(BuildContext context) {
    if (!kIsWeb) {
      return child;
    }

    return LayoutBuilder(
      builder: (context, constraints) {
        // If viewport is already mobile-sized, render full screen directly
        if (constraints.maxWidth <= 500) {
          return child;
        }

        // On desktop: Tall, responsive mobile layout (longer screen view, 1:1 fast rendering)
        const mobileWidth = 425.0;

        return Scaffold(
          backgroundColor: const Color(0xFF090D16),
          body: Center(
            child: Container(
              width: mobileWidth,
              height: double.infinity, // Full vertical height for longer screen view
              decoration: const BoxDecoration(
                color: Colors.white,
                border: Border(
                  left: BorderSide(color: Color(0xFF1E293B), width: 1.5),
                  right: BorderSide(color: Color(0xFF1E293B), width: 1.5),
                ),
                boxShadow: [
                  BoxShadow(
                    color: Color(0x66000000),
                    blurRadius: 24,
                    spreadRadius: 2,
                  ),
                ],
              ),
              child: ClipRect(
                child: _MobileScreenContent(child: child),
              ),
            ),
          ),
        );
      },
    );
  }
}

class _MobileScreenContent extends StatelessWidget {
  final Widget child;

  const _MobileScreenContent({required this.child});

  @override
  Widget build(BuildContext context) {
    final originalMq = MediaQuery.of(context);

    // Inject safe area padding: 38px top, 16px bottom
    // Ensures all app headers, titles, and buttons remain completely clear of the camera notch
    final customMq = originalMq.copyWith(
      padding: const EdgeInsets.only(top: 38, bottom: 16),
      viewPadding: const EdgeInsets.only(top: 38, bottom: 16),
      viewInsets: originalMq.viewInsets,
    );

    return MediaQuery(
      data: customMq,
      child: Stack(
        fit: StackFit.expand,
        children: [
          // 1. Mobile Application View (Full height, native 1:1 crisp rendering)
          child,

          // 2. Minimalist Mobile Camera Notch / Island (Ignored for touches)
          Positioned(
            top: 0,
            left: 0,
            right: 0,
            height: 38,
            child: IgnorePointer(
              child: Center(
                child: Container(
                  width: 108,
                  height: 24,
                  margin: const EdgeInsets.only(top: 6),
                  decoration: BoxDecoration(
                    color: Colors.black,
                    borderRadius: BorderRadius.circular(12),
                    border: Border.all(
                      color: const Color(0xFF1E293B),
                      width: 0.8,
                    ),
                  ),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      // Camera Lens
                      Container(
                        width: 8,
                        height: 8,
                        decoration: BoxDecoration(
                          shape: BoxShape.circle,
                          color: const Color(0xFF0F172A),
                          border: Border.all(color: const Color(0xFF334155), width: 0.8),
                        ),
                        child: Center(
                          child: Container(
                            width: 2.5,
                            height: 2.5,
                            decoration: const BoxDecoration(
                              color: Color(0xFF38BDF8),
                              shape: BoxShape.circle,
                            ),
                          ),
                        ),
                      ),
                      const SizedBox(width: 10),
                      // Sensor
                      Container(
                        width: 4,
                        height: 4,
                        decoration: const BoxDecoration(
                          color: Color(0xFF1E293B),
                          shape: BoxShape.circle,
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ),

          // 3. Minimal Home Indicator Bar at Bottom
          Positioned(
            bottom: 4,
            left: 0,
            right: 0,
            child: IgnorePointer(
              child: Center(
                child: Container(
                  width: 120,
                  height: 4,
                  decoration: BoxDecoration(
                    color: Colors.black.withValues(alpha: 0.25),
                    borderRadius: BorderRadius.circular(2),
                  ),
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }
}
