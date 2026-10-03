import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/bouncy_button.dart';
import '../../../core/widgets/glass_card.dart';
import '../../../core/widgets/gradient_button.dart';
import '../../../core/widgets/section_header.dart';
import '../models/conversion_models.dart';
import '../services/conversion_service.dart';

class RailwayTab extends StatefulWidget {
  const RailwayTab({super.key});

  @override
  State<RailwayTab> createState() => _RailwayTabState();
}

class _RailwayTabState extends State<RailwayTab> {
  final ConversionService _service = ConversionService();

  String _direction = '12to24'; // 12to24 | 24to12

  // 12-to-24 inputs
  int _hour12 = 8;
  int _minute12 = 30;
  String _ampm = 'PM';

  // 24-to-12 inputs
  int _hour24 = 20;
  int _minute24 = 30;

  bool _isLoading = false;
  RailwayResult? _result;
  String? _errorMessage;

  Future<void> _convert() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      RailwayResult res;
      if (_direction == '12to24') {
        res = await _service.convertRailway(
          direction: '12to24',
          hour: _hour12,
          minute: _minute12,
          ampm: _ampm,
        );
      } else {
        res = await _service.convertRailway(
          direction: '24to12',
          hour: _hour24,
          minute: _minute24,
        );
      }

      setState(() {
        _isLoading = false;
        _result = res;
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
            title: 'Railway Time Converter',
            subtitle: 'Switch between Standard 12-Hour AM/PM and 24-Hour Military/Railway Time',
            icon: Icons.train_rounded,
          ),
          const SizedBox(height: 12),

          // Direction Toggle
          Row(
            children: [
              Expanded(
                child: BouncyButton(
                  onTap: () => setState(() {
                    _direction = '12to24';
                    _result = null;
                  }),
                  child: Container(
                    padding: const EdgeInsets.symmetric(vertical: 12),
                    decoration: BoxDecoration(
                      color: _direction == '12to24' ? AppTheme.primary : Colors.white,
                      borderRadius: BorderRadius.circular(12),
                      border: Border.all(
                        color: _direction == '12to24' ? AppTheme.primary : const Color(0xFFE2E8F0),
                      ),
                      boxShadow: _direction == '12to24' ? AppTheme.softShadow : null,
                    ),
                    alignment: Alignment.center,
                    child: Text(
                      '12h AM/PM → 24h',
                      style: GoogleFonts.plusJakartaSans(
                        fontSize: 13,
                        fontWeight: FontWeight.w700,
                        color: _direction == '12to24' ? Colors.white : AppTheme.textSecondary,
                      ),
                    ),
                  ),
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: BouncyButton(
                  onTap: () => setState(() {
                    _direction = '24to12';
                    _result = null;
                  }),
                  child: Container(
                    padding: const EdgeInsets.symmetric(vertical: 12),
                    decoration: BoxDecoration(
                      color: _direction == '24to12' ? AppTheme.primary : Colors.white,
                      borderRadius: BorderRadius.circular(12),
                      border: Border.all(
                        color: _direction == '24to12' ? AppTheme.primary : const Color(0xFFE2E8F0),
                      ),
                      boxShadow: _direction == '24to12' ? AppTheme.softShadow : null,
                    ),
                    alignment: Alignment.center,
                    child: Text(
                      '24h → 12h AM/PM',
                      style: GoogleFonts.plusJakartaSans(
                        fontSize: 13,
                        fontWeight: FontWeight.w700,
                        color: _direction == '24to12' ? Colors.white : AppTheme.textSecondary,
                      ),
                    ),
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),

          GlassCard(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                if (_direction == '12to24') ...[
                  Row(
                    children: [
                      // Hour
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text('Hour', style: GoogleFonts.plusJakartaSans(fontSize: 12, fontWeight: FontWeight.w600)),
                            const SizedBox(height: 6),
                            Container(
                              padding: const EdgeInsets.symmetric(horizontal: 12),
                              decoration: BoxDecoration(
                                color: Colors.white,
                                borderRadius: BorderRadius.circular(12),
                                border: Border.all(color: const Color(0xFFE2E8F0)),
                              ),
                              child: DropdownButtonHideUnderline(
                                child: DropdownButton<int>(
                                  value: _hour12,
                                  isExpanded: true,
                                  items: List.generate(12, (i) => i + 1).map((h) {
                                    return DropdownMenuItem(value: h, child: Text(h.toString().padLeft(2, '0')));
                                  }).toList(),
                                  onChanged: (v) => setState(() => _hour12 = v ?? 1),
                                ),
                              ),
                            ),
                          ],
                        ),
                      ),
                      const SizedBox(width: 8),
                      // Minute
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text('Minute', style: GoogleFonts.plusJakartaSans(fontSize: 12, fontWeight: FontWeight.w600)),
                            const SizedBox(height: 6),
                            Container(
                              padding: const EdgeInsets.symmetric(horizontal: 12),
                              decoration: BoxDecoration(
                                color: Colors.white,
                                borderRadius: BorderRadius.circular(12),
                                border: Border.all(color: const Color(0xFFE2E8F0)),
                              ),
                              child: DropdownButtonHideUnderline(
                                child: DropdownButton<int>(
                                  value: _minute12,
                                  isExpanded: true,
                                  items: List.generate(60, (i) => i).map((m) {
                                    return DropdownMenuItem(value: m, child: Text(m.toString().padLeft(2, '0')));
                                  }).toList(),
                                  onChanged: (v) => setState(() => _minute12 = v ?? 0),
                                ),
                              ),
                            ),
                          ],
                        ),
                      ),
                      const SizedBox(width: 8),
                      // AM/PM
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text('AM/PM', style: GoogleFonts.plusJakartaSans(fontSize: 12, fontWeight: FontWeight.w600)),
                            const SizedBox(height: 6),
                            Container(
                              padding: const EdgeInsets.symmetric(horizontal: 12),
                              decoration: BoxDecoration(
                                color: Colors.white,
                                borderRadius: BorderRadius.circular(12),
                                border: Border.all(color: const Color(0xFFE2E8F0)),
                              ),
                              child: DropdownButtonHideUnderline(
                                child: DropdownButton<String>(
                                  value: _ampm,
                                  isExpanded: true,
                                  items: ['AM', 'PM'].map((ap) {
                                    return DropdownMenuItem(value: ap, child: Text(ap));
                                  }).toList(),
                                  onChanged: (v) => setState(() => _ampm = v ?? 'AM'),
                                ),
                              ),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                ] else ...[
                  Row(
                    children: [
                      // 24 Hour
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text('24h Hour', style: GoogleFonts.plusJakartaSans(fontSize: 12, fontWeight: FontWeight.w600)),
                            const SizedBox(height: 6),
                            Container(
                              padding: const EdgeInsets.symmetric(horizontal: 12),
                              decoration: BoxDecoration(
                                color: Colors.white,
                                borderRadius: BorderRadius.circular(12),
                                border: Border.all(color: const Color(0xFFE2E8F0)),
                              ),
                              child: DropdownButtonHideUnderline(
                                child: DropdownButton<int>(
                                  value: _hour24,
                                  isExpanded: true,
                                  items: List.generate(24, (i) => i).map((h) {
                                    return DropdownMenuItem(value: h, child: Text(h.toString().padLeft(2, '0')));
                                  }).toList(),
                                  onChanged: (v) => setState(() => _hour24 = v ?? 0),
                                ),
                              ),
                            ),
                          ],
                        ),
                      ),
                      const SizedBox(width: 12),
                      // 24 Minute
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text('Minute', style: GoogleFonts.plusJakartaSans(fontSize: 12, fontWeight: FontWeight.w600)),
                            const SizedBox(height: 6),
                            Container(
                              padding: const EdgeInsets.symmetric(horizontal: 12),
                              decoration: BoxDecoration(
                                color: Colors.white,
                                borderRadius: BorderRadius.circular(12),
                                border: Border.all(color: const Color(0xFFE2E8F0)),
                              ),
                              child: DropdownButtonHideUnderline(
                                child: DropdownButton<int>(
                                  value: _minute24,
                                  isExpanded: true,
                                  items: List.generate(60, (i) => i).map((m) {
                                    return DropdownMenuItem(value: m, child: Text(m.toString().padLeft(2, '0')));
                                  }).toList(),
                                  onChanged: (v) => setState(() => _minute24 = v ?? 0),
                                ),
                              ),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                ],
                const SizedBox(height: 24),
                GradientButton(
                  text: 'Convert Railway Time',
                  isLoading: _isLoading,
                  icon: Icons.schedule_send_rounded,
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
              padding: const EdgeInsets.all(24),
              child: Column(
                children: [
                  Text(
                    _direction == '12to24' ? '24-Hour Railway Time' : '12-Hour Standard Time',
                    style: GoogleFonts.plusJakartaSans(fontSize: 13, fontWeight: FontWeight.w600, color: AppTheme.textSecondary),
                  ),
                  const SizedBox(height: 8),
                  Text(
                    _direction == '12to24'
                        ? '${_result!.hour24?.toString().padLeft(2, '0')}:${_result!.minute.toString().padLeft(2, '0')} hrs'
                        : '${_result!.hour12?.toString().padLeft(2, '0')}:${_result!.minute.toString().padLeft(2, '0')} ${_result!.ampm ?? ''}',
                    style: GoogleFonts.plusJakartaSans(
                      fontSize: 32,
                      fontWeight: FontWeight.w800,
                      color: AppTheme.primary,
                      letterSpacing: -0.5,
                    ),
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
