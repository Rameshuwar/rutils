import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:intl/intl.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/bouncy_button.dart';
import '../../../core/widgets/glass_card.dart';
import '../../../core/widgets/gradient_button.dart';
import '../../../core/widgets/section_header.dart';
import '../models/calculation_models.dart';
import '../services/calculation_service.dart';

class AgeCalculatorTab extends StatefulWidget {
  const AgeCalculatorTab({super.key});

  @override
  State<AgeCalculatorTab> createState() => _AgeCalculatorTabState();
}

class _AgeCalculatorTabState extends State<AgeCalculatorTab> {
  final CalculationService _service = CalculationService();

  DateTime _dob = DateTime(1995, 6, 15);
  DateTime _today = DateTime.now();
  bool _isLoading = false;
  AgeResultModel? _result;
  String? _errorMessage;

  Future<void> _pickDob() async {
    final picked = await showDatePicker(
      context: context,
      initialDate: _dob,
      firstDate: DateTime(1900),
      lastDate: DateTime.now(),
    );
    if (picked != null) setState(() => _dob = picked);
  }

  Future<void> _pickToday() async {
    final picked = await showDatePicker(
      context: context,
      initialDate: _today,
      firstDate: DateTime(1900),
      lastDate: DateTime(2100),
    );
    if (picked != null) setState(() => _today = picked);
  }

  Future<void> _calculate() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      final dobStr = DateFormat('yyyy-MM-dd').format(_dob);
      final todayStr = DateFormat('yyyy-MM-dd').format(_today);

      final res = await _service.calculateAge(dob: dobStr, today: todayStr);
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
            title: 'Age Calculator',
            subtitle: 'Calculate exact chronological age, next birthday countdown & lifespan statistics',
            icon: Icons.cake_rounded,
          ),
          const SizedBox(height: 12),
          GlassCard(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text('Date of Birth', style: GoogleFonts.plusJakartaSans(fontSize: 12, fontWeight: FontWeight.w600)),
                          const SizedBox(height: 6),
                          BouncyButton(
                            onTap: _pickDob,
                            child: Container(
                              padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
                              decoration: BoxDecoration(
                                color: Colors.white,
                                borderRadius: BorderRadius.circular(12),
                                border: Border.all(color: const Color(0xFFE2E8F0)),
                              ),
                              child: Row(
                                children: [
                                  const Icon(Icons.cake_outlined, size: 16, color: AppTheme.primary),
                                  const SizedBox(width: 8),
                                  Text(DateFormat('dd MMM yyyy').format(_dob), style: GoogleFonts.plusJakartaSans(fontSize: 13, fontWeight: FontWeight.w600)),
                                ],
                              ),
                            ),
                          ),
                        ],
                      ),
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text('Age at Date', style: GoogleFonts.plusJakartaSans(fontSize: 12, fontWeight: FontWeight.w600)),
                          const SizedBox(height: 6),
                          BouncyButton(
                            onTap: _pickToday,
                            child: Container(
                              padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
                              decoration: BoxDecoration(
                                color: Colors.white,
                                borderRadius: BorderRadius.circular(12),
                                border: Border.all(color: const Color(0xFFE2E8F0)),
                              ),
                              child: Row(
                                children: [
                                  const Icon(Icons.calendar_today_rounded, size: 16, color: AppTheme.primary),
                                  const SizedBox(width: 8),
                                  Text(DateFormat('dd MMM yyyy').format(_today), style: GoogleFonts.plusJakartaSans(fontSize: 13, fontWeight: FontWeight.w600)),
                                ],
                              ),
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 24),
                GradientButton(
                  text: 'Calculate Age',
                  isLoading: _isLoading,
                  icon: Icons.calculate_rounded,
                  onPressed: _calculate,
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
                  Text('Current Age', style: GoogleFonts.plusJakartaSans(fontSize: 13, fontWeight: FontWeight.w600, color: AppTheme.textSecondary)),
                  const SizedBox(height: 8),
                  Text(
                    '${_result!.years} Years',
                    style: GoogleFonts.plusJakartaSans(
                      fontSize: 32,
                      fontWeight: FontWeight.w800,
                      color: AppTheme.primary,
                      letterSpacing: -0.5,
                    ),
                  ),
                  Text(
                    '${_result!.months} Months, ${_result!.days} Days',
                    style: GoogleFonts.plusJakartaSans(fontSize: 15, fontWeight: FontWeight.w600, color: AppTheme.textSecondary),
                  ),
                  const SizedBox(height: 16),
                  const Divider(color: Color(0xFFBFDBFE)),
                  const SizedBox(height: 12),

                  // Next Birthday
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text('Next Birthday', style: GoogleFonts.plusJakartaSans(fontSize: 11, color: AppTheme.textMuted)),
                          Text(
                            '${_result!.nextBirthday.dayOfWeek}, ${_result!.nextBirthday.date}',
                            style: GoogleFonts.plusJakartaSans(fontSize: 13, fontWeight: FontWeight.w700),
                          ),
                        ],
                      ),
                      Container(
                        padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
                        decoration: BoxDecoration(
                          color: AppTheme.primary,
                          borderRadius: BorderRadius.circular(10),
                        ),
                        child: Text(
                          '${_result!.nextBirthday.daysRemaining} days left',
                          style: GoogleFonts.plusJakartaSans(fontSize: 12, fontWeight: FontWeight.w700, color: Colors.white),
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 16),

                  // Summary Grid
                  Container(
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      color: Colors.white,
                      borderRadius: BorderRadius.circular(14),
                    ),
                    child: Row(
                      mainAxisAlignment: MainAxisAlignment.spaceAround,
                      children: [
                        _buildStat('Months', _result!.summary.months.toString()),
                        _buildStat('Weeks', _result!.summary.weeks.toString()),
                        _buildStat('Days', _result!.summary.days.toString()),
                        _buildStat('Hours', _result!.summary.hours.toString()),
                      ],
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

  Widget _buildStat(String label, String value) {
    return Column(
      children: [
        Text(value, style: GoogleFonts.plusJakartaSans(fontSize: 14, fontWeight: FontWeight.w800, color: AppTheme.textPrimary)),
        Text(label, style: GoogleFonts.plusJakartaSans(fontSize: 11, color: AppTheme.textMuted)),
      ],
    );
  }
}
