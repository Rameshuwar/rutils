import 'package:fl_chart/fl_chart.dart';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/utils/formatters.dart';
import '../../../core/widgets/app_dropdown.dart';
import '../../../core/widgets/bouncy_button.dart';
import '../../../core/widgets/custom_text_field.dart';
import '../../../core/widgets/glass_card.dart';
import '../../../core/widgets/gradient_button.dart';
import '../../../core/widgets/section_header.dart';
import '../models/calculation_models.dart';
import '../services/calculation_service.dart';

class EmiCalculatorTab extends StatefulWidget {
  const EmiCalculatorTab({super.key});

  @override
  State<EmiCalculatorTab> createState() => _EmiCalculatorTabState();
}

class _EmiCalculatorTabState extends State<EmiCalculatorTab> {
  final CalculationService _service = CalculationService();

  final _principalController = TextEditingController(text: '1000000');
  final _rateController = TextEditingController(text: '8.5');
  final _tenureController = TextEditingController(text: '5');

  String _tenureUnit = 'years'; // years | months
  bool _isLoading = false;
  EMIResultModel? _result;
  String? _errorMessage;
  bool _showAmortization = false;

  @override
  void dispose() {
    _principalController.dispose();
    _rateController.dispose();
    _tenureController.dispose();
    super.dispose();
  }

  Future<void> _calculate() async {
    final p = double.tryParse(_principalController.text);
    final r = double.tryParse(_rateController.text);
    final t = double.tryParse(_tenureController.text);

    if (p == null || p <= 0 || r == null || r < 0 || t == null || t <= 0) {
      setState(() => _errorMessage = 'Please enter valid loan details');
      return;
    }

    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      final res = await _service.calculateEMI(
        principal: p,
        annualInterestRate: r,
        tenure: t,
        tenureUnit: _tenureUnit,
      );

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
            title: 'Loan EMI Calculator',
            subtitle: 'Calculate monthly payments, interest vs principal breakdown & amortization',
            icon: Icons.account_balance_rounded,
          ),
          const SizedBox(height: 12),
          GlassCard(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                CustomTextField(
                  controller: _principalController,
                  label: 'Principal Loan Amount (₹)',
                  hint: 'e.g. 10,00,000',
                  keyboardType: const TextInputType.numberWithOptions(decimal: true),
                ),
                const SizedBox(height: 14),
                CustomTextField(
                  controller: _rateController,
                  label: 'Annual Interest Rate (%)',
                  hint: 'e.g. 8.5',
                  keyboardType: const TextInputType.numberWithOptions(decimal: true),
                ),
                const SizedBox(height: 14),
                Row(
                  children: [
                    Expanded(
                      flex: 3,
                      child: CustomTextField(
                        controller: _tenureController,
                        label: 'Loan Tenure',
                        hint: 'e.g. 5',
                        keyboardType: const TextInputType.numberWithOptions(decimal: true),
                      ),
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      flex: 2,
                      child: AppDropdown<String>(
                        label: 'Tenure Unit',
                        value: _tenureUnit,
                        items: const [
                          AppDropdownItem(value: 'years', label: 'Years', subtitle: 'Annual periods', icon: Icons.calendar_today_rounded),
                          AppDropdownItem(value: 'months', label: 'Months', subtitle: 'Monthly installments', icon: Icons.date_range_rounded),
                        ],
                        modalTitle: 'Select Loan Tenure Unit',
                        onChanged: (v) => setState(() => _tenureUnit = v),
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 24),
                GradientButton(
                  text: 'Calculate EMI',
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
                  Text('Monthly EMI Payable', style: GoogleFonts.plusJakartaSans(fontSize: 13, fontWeight: FontWeight.w600, color: AppTheme.textSecondary)),
                  const SizedBox(height: 6),
                  Text(
                    Formatters.inr(_result!.emi),
                    style: GoogleFonts.plusJakartaSans(
                      fontSize: 34,
                      fontWeight: FontWeight.w800,
                      color: AppTheme.primary,
                      letterSpacing: -0.5,
                    ),
                  ),
                  const SizedBox(height: 18),

                  // Pie Chart
                  SizedBox(
                    height: 140,
                    child: PieChart(
                      PieChartData(
                        sectionsSpace: 3,
                        centerSpaceRadius: 36,
                        sections: [
                          PieChartSectionData(
                            value: _result!.principalPercent,
                            title: '${_result!.principalPercent.toStringAsFixed(1)}%',
                            color: AppTheme.primary,
                            radius: 34,
                            titleStyle: GoogleFonts.plusJakartaSans(fontSize: 11, fontWeight: FontWeight.bold, color: Colors.white),
                          ),
                          PieChartSectionData(
                            value: _result!.interestPercent,
                            title: '${_result!.interestPercent.toStringAsFixed(1)}%',
                            color: const Color(0xFF0EA5E9),
                            radius: 34,
                            titleStyle: GoogleFonts.plusJakartaSans(fontSize: 11, fontWeight: FontWeight.bold, color: Colors.white),
                          ),
                        ],
                      ),
                    ),
                  ),
                  const SizedBox(height: 16),

                  // Stat Cards
                  Row(
                    children: [
                      Expanded(
                        child: Container(
                          padding: const EdgeInsets.all(12),
                          decoration: BoxDecoration(color: Colors.white, borderRadius: BorderRadius.circular(12)),
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Row(
                                children: [
                                  Container(width: 8, height: 8, decoration: const BoxDecoration(color: AppTheme.primary, shape: BoxShape.circle)),
                                  const SizedBox(width: 6),
                                  Text('Principal', style: GoogleFonts.plusJakartaSans(fontSize: 11, color: AppTheme.textMuted)),
                                ],
                              ),
                              const SizedBox(height: 4),
                              Text(Formatters.inr(_result!.principal), style: GoogleFonts.plusJakartaSans(fontSize: 14, fontWeight: FontWeight.w700)),
                            ],
                          ),
                        ),
                      ),
                      const SizedBox(width: 10),
                      Expanded(
                        child: Container(
                          padding: const EdgeInsets.all(12),
                          decoration: BoxDecoration(color: Colors.white, borderRadius: BorderRadius.circular(12)),
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Row(
                                children: [
                                  Container(width: 8, height: 8, decoration: const BoxDecoration(color: Color(0xFF0EA5E9), shape: BoxShape.circle)),
                                  const SizedBox(width: 6),
                                  Text('Total Interest', style: GoogleFonts.plusJakartaSans(fontSize: 11, color: AppTheme.textMuted)),
                                ],
                              ),
                              const SizedBox(height: 4),
                              Text(Formatters.inr(_result!.totalInterest), style: GoogleFonts.plusJakartaSans(fontSize: 14, fontWeight: FontWeight.w700)),
                            ],
                          ),
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 10),
                  Container(
                    width: double.infinity,
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(color: Colors.white, borderRadius: BorderRadius.circular(12)),
                    child: Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Text('Total Amount Payable:', style: GoogleFonts.plusJakartaSans(fontSize: 13, color: AppTheme.textSecondary)),
                        Text(Formatters.inr(_result!.totalPayment), style: GoogleFonts.plusJakartaSans(fontSize: 15, fontWeight: FontWeight.w800, color: AppTheme.primary)),
                      ],
                    ),
                  ),

                  // Amortization Toggle
                  if (_result!.amortization.isNotEmpty) ...[
                    const SizedBox(height: 16),
                    BouncyButton(
                      onTap: () => setState(() => _showAmortization = !_showAmortization),
                      child: Container(
                        padding: const EdgeInsets.symmetric(vertical: 10, horizontal: 16),
                        decoration: BoxDecoration(
                          color: AppTheme.primary.withValues(alpha: 0.1),
                          borderRadius: BorderRadius.circular(12),
                        ),
                        child: Row(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            Icon(_showAmortization ? Icons.expand_less_rounded : Icons.expand_more_rounded, color: AppTheme.primary, size: 20),
                            const SizedBox(width: 8),
                            Text(
                              _showAmortization ? 'Hide Repayment Schedule' : 'View Repayment Schedule (${_result!.amortization.length} Months)',
                              style: GoogleFonts.plusJakartaSans(fontSize: 13, fontWeight: FontWeight.w700, color: AppTheme.primary),
                            ),
                          ],
                        ),
                      ),
                    ),
                    if (_showAmortization) ...[
                      const SizedBox(height: 12),
                      Container(
                        constraints: const BoxConstraints(maxHeight: 260),
                        decoration: BoxDecoration(
                          color: Colors.white,
                          borderRadius: BorderRadius.circular(12),
                          border: Border.all(color: const Color(0xFFE2E8F0)),
                        ),
                        child: ListView.separated(
                          itemCount: _result!.amortization.length,
                          separatorBuilder: (context, index) => const Divider(height: 1, color: Color(0xFFF1F5F9)),
                          itemBuilder: (context, idx) {
                            final row = _result!.amortization[idx];
                            return Padding(
                              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                              child: Row(
                                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                                children: [
                                  Text('Month ${row.month}', style: GoogleFonts.plusJakartaSans(fontSize: 12, fontWeight: FontWeight.w700)),
                                  Text('Principal: ${Formatters.inr(row.principalPaid, compact: true)}', style: GoogleFonts.plusJakartaSans(fontSize: 11, color: AppTheme.textSecondary)),
                                  Text('Interest: ${Formatters.inr(row.interestPaid, compact: true)}', style: GoogleFonts.plusJakartaSans(fontSize: 11, color: AppTheme.textMuted)),
                                  Text('Bal: ${Formatters.inr(row.closingBalance, compact: true)}', style: GoogleFonts.plusJakartaSans(fontSize: 11, fontWeight: FontWeight.w600)),
                                ],
                              ),
                            );
                          },
                        ),
                      ),
                    ],
                  ],
                ],
              ),
            ),
          ],
        ],
      ),
    );
  }
}
