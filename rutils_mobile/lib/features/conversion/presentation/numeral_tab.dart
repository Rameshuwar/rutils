import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:google_fonts/google_fonts.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/app_dropdown.dart';
import '../../../core/widgets/bouncy_button.dart';
import '../../../core/widgets/custom_text_field.dart';
import '../../../core/widgets/glass_card.dart';
import '../../../core/widgets/gradient_button.dart';
import '../../../core/widgets/section_header.dart';
import '../services/conversion_service.dart';

class NumeralTab extends StatefulWidget {
  const NumeralTab({super.key});

  @override
  State<NumeralTab> createState() => _NumeralTabState();
}

class _NumeralTabState extends State<NumeralTab> {
  final ConversionService _service = ConversionService();
  final _inputController = TextEditingController(text: '255');

  String _fromBase = 'decimal';
  String _toBase = 'binary';
  bool _isLoading = false;
  String? _result;
  String? _errorMessage;

  @override
  void dispose() {
    _inputController.dispose();
    super.dispose();
  }

  void _swap() {
    setState(() {
      final tmp = _fromBase;
      _fromBase = _toBase;
      _toBase = tmp;
      _result = null;
    });
  }

  Future<void> _convert() async {
    final val = _inputController.text.trim();
    if (val.isEmpty) {
      setState(() => _errorMessage = 'Please enter a number to convert');
      return;
    }

    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      final res = await _service.convertNumeral(
        value: val,
        fromBase: _fromBase,
        toBase: _toBase,
      );

      setState(() {
        _isLoading = false;
        _result = res.result;
      });
    } catch (e) {
      setState(() {
        _isLoading = false;
        _errorMessage = e.toString();
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      physics: const BouncingScrollPhysics(),
      padding: const EdgeInsets.all(20),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          const SectionHeader(
            title: 'Numeral System Converter',
            subtitle: 'Convert between Binary, Octal, Decimal, and Hexadecimal representations',
            icon: Icons.tag_rounded,
          ),
          const SizedBox(height: 12),
          GlassCard(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                CustomTextField(
                  controller: _inputController,
                  label: 'Input Value',
                  hint: _fromBase == 'binary'
                      ? 'e.g. 10101'
                      : _fromBase == 'hexadecimal'
                          ? 'e.g. FF'
                          : 'e.g. 255',
                ),
                const SizedBox(height: 16),
                AppDropdown<String>(
                  label: 'From Base System',
                  value: _fromBase,
                  items: const [
                    AppDropdownItem(value: 'decimal', label: 'Decimal (Base 10)', subtitle: '0-9 digits', icon: Icons.numbers_rounded),
                    AppDropdownItem(value: 'binary', label: 'Binary (Base 2)', subtitle: '0 and 1 bits', icon: Icons.memory_rounded),
                    AppDropdownItem(value: 'hexadecimal', label: 'Hexadecimal (Base 16)', subtitle: '0-9 and A-F', icon: Icons.tag_rounded),
                    AppDropdownItem(value: 'octal', label: 'Octal (Base 8)', subtitle: '0-7 digits', icon: Icons.pin_rounded),
                  ],
                  modalTitle: 'Select Source Base',
                  onChanged: (v) => setState(() => _fromBase = v),
                ),
                const SizedBox(height: 10),
                Center(
                  child: BouncyButton(
                    onTap: _swap,
                    child: Container(
                      padding: const EdgeInsets.all(8),
                      decoration: BoxDecoration(
                        color: const Color(0xFFEFF6FF),
                        borderRadius: BorderRadius.circular(10),
                        border: Border.all(color: const Color(0xFFBFDBFE)),
                      ),
                      child: const Icon(Icons.swap_vert_rounded, color: AppTheme.primary, size: 20),
                    ),
                  ),
                ),
                const SizedBox(height: 10),
                AppDropdown<String>(
                  label: 'To Base System',
                  value: _toBase,
                  items: const [
                    AppDropdownItem(value: 'binary', label: 'Binary (Base 2)', subtitle: '0 and 1 bits', icon: Icons.memory_rounded),
                    AppDropdownItem(value: 'decimal', label: 'Decimal (Base 10)', subtitle: '0-9 digits', icon: Icons.numbers_rounded),
                    AppDropdownItem(value: 'hexadecimal', label: 'Hexadecimal (Base 16)', subtitle: '0-9 and A-F', icon: Icons.tag_rounded),
                    AppDropdownItem(value: 'octal', label: 'Octal (Base 8)', subtitle: '0-7 digits', icon: Icons.pin_rounded),
                  ],
                  modalTitle: 'Select Target Base',
                  onChanged: (v) => setState(() => _toBase = v),
                ),
                const SizedBox(height: 24),
                GradientButton(
                  text: 'Convert Numeral',
                  isLoading: _isLoading,
                  icon: Icons.sync_rounded,
                  onPressed: _convert,
                ),
              ],
            ),
          ),
          if (_errorMessage != null) ...[
            const SizedBox(height: 16),
            Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: AppTheme.danger.withValues(alpha: 0.1),
                borderRadius: BorderRadius.circular(12),
              ),
              child: Text(_errorMessage!, style: GoogleFonts.plusJakartaSans(color: AppTheme.danger, fontSize: 13)),
            ),
          ],
          if (_result != null) ...[
            const SizedBox(height: 16),
            GlassCard(
              backgroundColor: const Color(0xFFEFF6FF),
              padding: const EdgeInsets.all(22),
              child: Column(
                children: [
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(
                        '${_toBase.toUpperCase()} RESULT',
                        style: GoogleFonts.plusJakartaSans(
                          fontSize: 12,
                          fontWeight: FontWeight.w700,
                          color: AppTheme.primary,
                        ),
                      ),
                      BouncyButton(
                        onTap: () {
                          Clipboard.setData(ClipboardData(text: _result!));
                          ScaffoldMessenger.of(context).showSnackBar(
                            const SnackBar(content: Text('Copied to clipboard!'), duration: Duration(seconds: 1)),
                          );
                        },
                        child: const Row(
                          children: [
                            Icon(Icons.copy_rounded, size: 14, color: AppTheme.primary),
                            SizedBox(width: 4),
                            Text('Copy', style: TextStyle(fontSize: 12, color: AppTheme.primary, fontWeight: FontWeight.bold)),
                          ],
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 12),
                  SelectableText(
                    _result!,
                    style: GoogleFonts.jetBrainsMono(
                      fontSize: 26,
                      fontWeight: FontWeight.w800,
                      color: AppTheme.textPrimary,
                    ),
                    textAlign: TextAlign.center,
                  ),
                ],
              ),
            ),
          ],
        ],
      ),
    );
  }
}
