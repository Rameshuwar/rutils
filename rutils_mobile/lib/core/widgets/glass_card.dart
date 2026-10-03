import 'dart:ui';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import '../theme/app_theme.dart';

class GlassCard extends StatelessWidget {
  final Widget child;
  final EdgeInsetsGeometry padding;
  final EdgeInsetsGeometry? margin;
  final double borderRadius;
  final Color? backgroundColor;
  final Border? border;
  final List<BoxShadow>? shadows;
  final double blurSigma;

  const GlassCard({
    super.key,
    required this.child,
    this.padding = const EdgeInsets.all(20),
    this.margin,
    this.borderRadius = 22,
    this.backgroundColor,
    this.border,
    this.shadows,
    this.blurSigma = 10,
  });

  @override
  Widget build(BuildContext context) {
    final effectiveShadows = shadows ?? AppTheme.softShadow;
    final effectiveBorder = border ??
        Border.all(
          color: Colors.white.withValues(alpha: 0.8),
          width: 1.2,
        );

    final cardContent = Container(
      padding: padding,
      decoration: BoxDecoration(
        color: backgroundColor ?? Colors.white.withValues(alpha: 0.94),
        borderRadius: BorderRadius.circular(borderRadius),
        border: effectiveBorder,
      ),
      child: child,
    );

    // On Web, skip expensive GPU BackdropFilter layer reading for silky-smooth 60fps
    if (kIsWeb || blurSigma <= 0) {
      return Container(
        margin: margin,
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(borderRadius),
          boxShadow: effectiveShadows,
        ),
        child: ClipRRect(
          borderRadius: BorderRadius.circular(borderRadius),
          child: cardContent,
        ),
      );
    }

    return Container(
      margin: margin,
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(borderRadius),
        boxShadow: effectiveShadows,
      ),
      child: ClipRRect(
        borderRadius: BorderRadius.circular(borderRadius),
        child: BackdropFilter(
          filter: ImageFilter.blur(sigmaX: blurSigma, sigmaY: blurSigma),
          child: cardContent,
        ),
      ),
    );
  }
}
