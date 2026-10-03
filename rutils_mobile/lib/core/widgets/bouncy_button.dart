import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

class BouncyButton extends StatefulWidget {
  final Widget child;
  final VoidCallback? onTap;
  final double scaleFactor;
  final Duration duration;

  const BouncyButton({
    super.key,
    required this.child,
    required this.onTap,
    this.scaleFactor = 0.96,
    this.duration = const Duration(milliseconds: 100),
  });

  @override
  State<BouncyButton> createState() => _BouncyButtonState();
}

class _BouncyButtonState extends State<BouncyButton> {
  bool _isDown = false;

  void _handleTapDown(TapDownDetails _) {
    if (widget.onTap != null) {
      setState(() => _isDown = true);
      HapticFeedback.selectionClick();
    }
  }

  void _handleTapUp(TapUpDetails _) {
    if (widget.onTap != null) {
      setState(() => _isDown = false);
    }
  }

  void _handleTapCancel() {
    if (widget.onTap != null) {
      setState(() => _isDown = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      behavior: HitTestBehavior.opaque,
      onTapDown: _handleTapDown,
      onTapUp: _handleTapUp,
      onTapCancel: _handleTapCancel,
      onTap: widget.onTap,
      child: AnimatedScale(
        scale: _isDown ? widget.scaleFactor : 1.0,
        duration: widget.duration,
        curve: Curves.easeOutCubic,
        child: widget.child,
      ),
    );
  }
}
