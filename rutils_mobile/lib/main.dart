import 'dart:ui';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'core/network/api_client.dart';
import 'core/theme/app_theme.dart';
import 'core/widgets/mobile_device_frame.dart';
import 'features/auth/presentation/login_screen.dart';
import 'features/auth/services/auth_service.dart';
import 'features/main_layout/presentation/main_nav_scaffold.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();

  // Edge to edge system overlay
  SystemChrome.setSystemUIOverlayStyle(
    const SystemUiOverlayStyle(
      statusBarColor: Colors.transparent,
      statusBarIconBrightness: Brightness.dark,
      statusBarBrightness: Brightness.light,
      systemNavigationBarColor: Colors.white,
      systemNavigationBarIconBrightness: Brightness.dark,
    ),
  );

  try {
    final apiClient = ApiClient();
    await apiClient.initFromStorage();

    final authService = AuthService();
    await authService.init();
  } catch (e, st) {
    debugPrint('Init error: $e\n$st');
  }

  runApp(const RutilsApp());
}

/// Allows smooth click-and-drag gestures on desktop browsers matching touch screens
class MobileScrollBehavior extends MaterialScrollBehavior {
  @override
  Set<PointerDeviceKind> get dragDevices => {
    PointerDeviceKind.touch,
    PointerDeviceKind.mouse,
    PointerDeviceKind.trackpad,
    PointerDeviceKind.stylus,
  };
}

class RutilsApp extends StatelessWidget {
  const RutilsApp({super.key});

  @override
  Widget build(BuildContext context) {
    return ListenableBuilder(
      listenable: AuthService(),
      builder: (context, _) {
        final isLoggedIn = AuthService().isLoggedIn;

        return MaterialApp(
          title: 'Rutils Mobile',
          debugShowCheckedModeBanner: false,
          theme: AppTheme.lightTheme,
          scrollBehavior: MobileScrollBehavior(),
          builder: (context, child) => MobileDeviceFrame(child: child ?? const SizedBox.shrink()),
          home: isLoggedIn ? const MainNavScaffold() : const LoginScreen(),
        );
      },
    );
  }
}
