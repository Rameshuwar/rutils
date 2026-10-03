import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/custom_text_field.dart';
import '../../../core/widgets/gradient_button.dart';
import '../services/auth_service.dart';

class ResetPasswordDialog extends StatefulWidget {
  final bool isForced;

  const ResetPasswordDialog({super.key, this.isForced = false});

  static Future<bool?> show(BuildContext context, {bool isForced = false}) {
    return showDialog<bool>(
      context: context,
      barrierDismissible: !isForced,
      builder: (context) => ResetPasswordDialog(isForced: isForced),
    );
  }

  @override
  State<ResetPasswordDialog> createState() => _ResetPasswordDialogState();
}

class _ResetPasswordDialogState extends State<ResetPasswordDialog> {
  final _newPasswordController = TextEditingController();
  final _confirmPasswordController = TextEditingController();
  bool _isLoading = false;
  String? _errorMessage;

  @override
  void dispose() {
    _newPasswordController.dispose();
    _confirmPasswordController.dispose();
    super.dispose();
  }

  Future<void> _handleReset() async {
    final newPass = _newPasswordController.text;
    final confirmPass = _confirmPasswordController.text;

    if (newPass.length < 8) {
      setState(() => _errorMessage = 'Password must be at least 8 characters long.');
      return;
    }
    if (newPass != confirmPass) {
      setState(() => _errorMessage = 'Passwords do not match.');
      return;
    }

    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      await AuthService().resetPassword(
        newPassword: newPass,
        confirmPassword: confirmPass,
      );
      if (mounted) {
        Navigator.of(context).pop(true);
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _errorMessage = e.toString();
          _isLoading = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Dialog(
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(24)),
      backgroundColor: Colors.white,
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Container(
                  padding: const EdgeInsets.all(10),
                  decoration: BoxDecoration(
                    color: AppTheme.warning.withValues(alpha: 0.12),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: const Icon(Icons.security_rounded, color: AppTheme.warning, size: 24),
                ),
                const SizedBox(width: 14),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        widget.isForced ? 'Set New Password' : 'Reset Password',
                        style: GoogleFonts.plusJakartaSans(
                          fontSize: 18,
                          fontWeight: FontWeight.w700,
                          color: AppTheme.textPrimary,
                        ),
                      ),
                      Text(
                        widget.isForced
                            ? 'Security requirement: temporary password used.'
                            : 'Enter your new strong password',
                        style: GoogleFonts.plusJakartaSans(
                          fontSize: 12,
                          color: AppTheme.textMuted,
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
            const SizedBox(height: 20),
            CustomTextField(
              controller: _newPasswordController,
              label: 'New Password',
              hint: 'Min 8 chars, uppercase, number & symbol',
              prefixIcon: Icons.lock_outline,
              obscureText: true,
            ),
            const SizedBox(height: 14),
            CustomTextField(
              controller: _confirmPasswordController,
              label: 'Confirm New Password',
              hint: 'Re-enter your new password',
              prefixIcon: Icons.lock_reset,
              obscureText: true,
            ),
            if (_errorMessage != null) ...[
              const SizedBox(height: 12),
              Text(
                _errorMessage!,
                style: GoogleFonts.plusJakartaSans(
                  fontSize: 12,
                  color: AppTheme.danger,
                  fontWeight: FontWeight.w500,
                ),
              ),
            ],
            const SizedBox(height: 24),
            Row(
              children: [
                if (!widget.isForced) ...[
                  Expanded(
                    child: OutlinedButton(
                      onPressed: () => Navigator.pop(context),
                      style: OutlinedButton.styleFrom(
                        padding: const EdgeInsets.symmetric(vertical: 14),
                        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
                      ),
                      child: Text('Cancel', style: GoogleFonts.plusJakartaSans(fontWeight: FontWeight.w600)),
                    ),
                  ),
                  const SizedBox(width: 12),
                ],
                Expanded(
                  child: GradientButton(
                    text: 'Save Password',
                    isLoading: _isLoading,
                    onPressed: _handleReset,
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}
