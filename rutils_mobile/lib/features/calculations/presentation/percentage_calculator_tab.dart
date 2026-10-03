import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/app_dropdown.dart';
import '../../../core/widgets/custom_text_field.dart';
import '../../../core/widgets/glass_card.dart';
import '../../../core/widgets/gradient_button.dart';
import '../../../core/widgets/section_header.dart';
import '../models/calculation_models.dart';
import '../services/calculation_service.dart';

class PercentageCalculatorTab extends StatefulWidget {
  const PercentageCalculatorTab({super.key});

  @override
  State<PercentageCalculatorTab> createState() => _PercentageCalculatorTabState();
}

class _PercentageCalculatorTabState extends State<PercentageCalculatorTab> {
  final CalculationService _service = CalculationService();

  final _field1Controller = TextEditingController(text: '15');
  final _field2Controller = TextEditingController(text: '200');
  final _field3Controller = TextEditingController(text: '12');

  String _operation = 'percent_of';
  bool _isLoading = false;
  PercentageResultModel? _result;
  String? _errorMessage;

  static const Map<String, ({
    String label,
    String subtitle,
    IconData icon,
    String f1Label,
    String f1Hint,
    String? f2Label,
    String? f2Hint,
    String? f3Label,
    String? f3Hint,
  })> _operations = {
    'percent_of': (
      label: 'What is X% of Y?',
      subtitle: 'Calculate standard percentage share',
      icon: Icons.percent_rounded,
      f1Label: 'Percentage (%)',
      f1Hint: 'e.g. 15',
      f2Label: 'Total / Base Number',
      f2Hint: 'e.g. 200',
      f3Label: null,
      f3Hint: null,
    ),
    'what_percent': (
      label: 'X is what % of Y?',
      subtitle: 'Find percentage ratio between part and whole',
      icon: Icons.pie_chart_outline_rounded,
      f1Label: 'Part Value (X)',
      f1Hint: 'e.g. 30',
      f2Label: 'Whole Value (Y)',
      f2Hint: 'e.g. 200',
      f3Label: null,
      f3Hint: null,
    ),
    'is_percent_of_what': (
      label: 'X is Y% of what?',
      subtitle: 'Find the total when part and percent are known',
      icon: Icons.help_outline_rounded,
      f1Label: 'Part Value (X)',
      f1Hint: 'e.g. 30',
      f2Label: 'Percentage (%)',
      f2Hint: 'e.g. 15',
      f3Label: null,
      f3Hint: null,
    ),
    'percent_change': (
      label: 'Percentage Change from X to Y',
      subtitle: 'Calculate relative increase or decrease rate',
      icon: Icons.trending_up_rounded,
      f1Label: 'Initial Value (Old)',
      f1Hint: 'e.g. 100',
      f2Label: 'Final Value (New)',
      f2Hint: 'e.g. 150',
      f3Label: null,
      f3Hint: null,
    ),
    'percent_increase': (
      label: 'Increase X by Y%',
      subtitle: 'Add percentage to base value',
      icon: Icons.arrow_upward_rounded,
      f1Label: 'Base Value',
      f1Hint: 'e.g. 200',
      f2Label: 'Increase By (%)',
      f2Hint: 'e.g. 15',
      f3Label: null,
      f3Hint: null,
    ),
    'percent_decrease': (
      label: 'Decrease X by Y%',
      subtitle: 'Deduct percentage from base value',
      icon: Icons.arrow_downward_rounded,
      f1Label: 'Base Value',
      f1Hint: 'e.g. 200',
      f2Label: 'Decrease By (%)',
      f2Hint: 'e.g. 15',
      f3Label: null,
      f3Hint: null,
    ),
    'reverse_percent': (
      label: 'Reverse Percentage',
      subtitle: 'Recover original price before tax / percentage was applied',
      icon: Icons.history_rounded,
      f1Label: 'Final / Current Value',
      f1Hint: 'e.g. 118',
      f2Label: 'Applied Percentage (%)',
      f2Hint: 'e.g. 18',
      f3Label: null,
      f3Hint: null,
    ),
    'percent_difference': (
      label: 'Percentage Difference',
      subtitle: 'Measure relative difference between two values',
      icon: Icons.compare_arrows_rounded,
      f1Label: 'First Value (A)',
      f1Hint: 'e.g. 80',
      f2Label: 'Second Value (B)',
      f2Hint: 'e.g. 100',
      f3Label: null,
      f3Hint: null,
    ),
    'discount': (
      label: 'Discount & Sale Price',
      subtitle: 'Calculate savings and final discounted amount',
      icon: Icons.local_offer_outlined,
      f1Label: 'Original Price (₹)',
      f1Hint: 'e.g. 1500',
      f2Label: 'Discount Rate (%)',
      f2Hint: 'e.g. 25',
      f3Label: null,
      f3Hint: null,
    ),
    'markup': (
      label: 'Markup & Retail Price',
      subtitle: 'Calculate cost markup and retail price',
      icon: Icons.storefront_rounded,
      f1Label: 'Cost Price (₹)',
      f1Hint: 'e.g. 800',
      f2Label: 'Markup Rate (%)',
      f2Hint: 'e.g. 35',
      f3Label: null,
      f3Hint: null,
    ),
    'profit_loss': (
      label: 'Profit / Loss Percentage',
      subtitle: 'Calculate profit or loss margins from cost & sale',
      icon: Icons.account_balance_wallet_rounded,
      f1Label: 'Cost Price (₹)',
      f1Hint: 'e.g. 500',
      f2Label: 'Selling Price (₹)',
      f2Hint: 'e.g. 650',
      f3Label: null,
      f3Hint: null,
    ),
    'compound_percent': (
      label: 'Compound Percentage Growth',
      subtitle: 'Calculate exponential compounded rate across periods',
      icon: Icons.auto_graph_rounded,
      f1Label: 'Initial Principal / Value',
      f1Hint: 'e.g. 10000',
      f2Label: 'Rate per Period (%)',
      f2Hint: 'e.g. 5',
      f3Label: 'Number of Periods',
      f3Hint: 'e.g. 12',
    ),
    'marks_percentage': (
      label: 'Exam Marks Percentage',
      subtitle: 'Calculate academic percentage from score and total',
      icon: Icons.school_rounded,
      f1Label: 'Marks Obtained',
      f1Hint: 'e.g. 465',
      f2Label: 'Maximum Total Marks',
      f2Hint: 'e.g. 500',
      f3Label: null,
      f3Hint: null,
    ),
    'cgpa_to_percent': (
      label: 'CGPA to Percentage (CBSE 9.5x)',
      subtitle: 'Convert 10-point CGPA to equivalent percentage',
      icon: Icons.grade_rounded,
      f1Label: 'CGPA Score (0 - 10)',
      f1Hint: 'e.g. 8.4',
      f2Label: null,
      f2Hint: null,
      f3Label: null,
      f3Hint: null,
    ),
    'add_percent_points': (
      label: 'Add Percentage Points',
      subtitle: 'Direct percentage point addition (P1 + P2)',
      icon: Icons.add_circle_outline_rounded,
      f1Label: 'First Rate (%)',
      f1Hint: 'e.g. 5.5',
      f2Label: 'Points to Add (%)',
      f2Hint: 'e.g. 1.5',
      f3Label: null,
      f3Hint: null,
    ),
    'subtract_percent_points': (
      label: 'Subtract Percentage Points',
      subtitle: 'Direct percentage point deduction (P1 - P2)',
      icon: Icons.remove_circle_outline_rounded,
      f1Label: 'Starting Rate (%)',
      f1Hint: 'e.g. 8.0',
      f2Label: 'Points to Deduct (%)',
      f2Hint: 'e.g. 2.5',
      f3Label: null,
      f3Hint: null,
    ),
    'fraction_to_percent': (
      label: 'Fraction to Percentage',
      subtitle: 'Convert fraction (Numerator / Denominator) to %',
      icon: Icons.functions_rounded,
      f1Label: 'Numerator (Top)',
      f1Hint: 'e.g. 3',
      f2Label: 'Denominator (Bottom)',
      f2Hint: 'e.g. 4',
      f3Label: null,
      f3Hint: null,
    ),
    'percent_to_fraction': (
      label: 'Percentage to Fraction',
      subtitle: 'Convert percentage to lowest simplified fraction',
      icon: Icons.exposure_rounded,
      f1Label: 'Percentage (%)',
      f1Hint: 'e.g. 75',
      f2Label: null,
      f2Hint: null,
      f3Label: null,
      f3Hint: null,
    ),
    'decimal_to_percent': (
      label: 'Decimal to Percentage',
      subtitle: 'Multiply decimal number by 100 to get percentage',
      icon: Icons.calculate_outlined,
      f1Label: 'Decimal Number',
      f1Hint: 'e.g. 0.375',
      f2Label: null,
      f2Hint: null,
      f3Label: null,
      f3Hint: null,
    ),
  };

  @override
  void dispose() {
    _field1Controller.dispose();
    _field2Controller.dispose();
    _field3Controller.dispose();
    super.dispose();
  }

  Future<void> _calculate() async {
    final config = _operations[_operation]!;
    final v1 = double.tryParse(_field1Controller.text.trim());
    final v2 = config.f2Label != null ? double.tryParse(_field2Controller.text.trim()) : null;
    final v3 = config.f3Label != null ? double.tryParse(_field3Controller.text.trim()) : null;

    if (v1 == null) {
      setState(() => _errorMessage = 'Please enter a valid number for ${config.f1Label}');
      return;
    }
    if (config.f2Label != null && v2 == null) {
      setState(() => _errorMessage = 'Please enter a valid number for ${config.f2Label}');
      return;
    }
    if (config.f3Label != null && v3 == null) {
      setState(() => _errorMessage = 'Please enter a valid number for ${config.f3Label}');
      return;
    }

    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      final res = await _service.calculatePercentage(
        operation: _operation,
        value1: v1,
        value2: v2,
        value3: v3,
      );

      setState(() {
        _isLoading = false;
        _result = res;
      });
    } catch (e) {
      setState(() {
        _isLoading = false;
        _errorMessage = e.toString().replaceFirst('Exception: ', '');
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    final config = _operations[_operation]!;

    final dropdownItems = _operations.entries.map((entry) {
      return AppDropdownItem<String>(
        value: entry.key,
        label: entry.value.label,
        subtitle: entry.value.subtitle,
        icon: entry.value.icon,
      );
    }).toList();

    return SingleChildScrollView(
      physics: const BouncingScrollPhysics(),
      padding: const EdgeInsets.all(20),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          const SectionHeader(
            title: 'Percentage Suite',
            subtitle: '19 specialized percentage, finance, academic & growth calculators',
            icon: Icons.percent_rounded,
          ),
          const SizedBox(height: 12),
          GlassCard(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // Modern Improvised Dropdown
                AppDropdown<String>(
                  label: 'Calculation Type',
                  value: _operation,
                  items: dropdownItems,
                  modalTitle: 'Choose Percentage Calculator',
                  enableSearch: true,
                  onChanged: (val) {
                    setState(() {
                      _operation = val;
                      _result = null;
                      _errorMessage = null;
                    });
                  },
                ),
                const SizedBox(height: 16),

                // Field 1
                CustomTextField(
                  controller: _field1Controller,
                  label: config.f1Label,
                  hint: config.f1Hint,
                  keyboardType: const TextInputType.numberWithOptions(decimal: true),
                ),

                // Field 2 (if required)
                if (config.f2Label != null) ...[
                  const SizedBox(height: 14),
                  CustomTextField(
                    controller: _field2Controller,
                    label: config.f2Label!,
                    hint: config.f2Hint!,
                    keyboardType: const TextInputType.numberWithOptions(decimal: true),
                  ),
                ],

                // Field 3 (for compound growth)
                if (config.f3Label != null) ...[
                  const SizedBox(height: 14),
                  CustomTextField(
                    controller: _field3Controller,
                    label: config.f3Label!,
                    hint: config.f3Hint!,
                    keyboardType: const TextInputType.numberWithOptions(decimal: true),
                  ),
                ],

                const SizedBox(height: 24),
                GradientButton(
                  text: 'Calculate Result',
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
              padding: const EdgeInsets.all(14),
              decoration: BoxDecoration(
                color: AppTheme.danger.withValues(alpha: 0.1),
                borderRadius: BorderRadius.circular(14),
                border: Border.all(color: AppTheme.danger.withValues(alpha: 0.3)),
              ),
              child: Row(
                children: [
                  const Icon(Icons.error_outline_rounded, color: AppTheme.danger, size: 20),
                  const SizedBox(width: 10),
                  Expanded(
                    child: Text(
                      _errorMessage!,
                      style: GoogleFonts.plusJakartaSans(
                        color: AppTheme.danger,
                        fontSize: 13,
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ],
          if (_result != null) ...[
            const SizedBox(height: 16),
            GlassCard(
              backgroundColor: const Color(0xFFEFF6FF),
              padding: const EdgeInsets.all(22),
              child: Column(
                children: [
                  Text(
                    'Calculated Result',
                    style: GoogleFonts.plusJakartaSans(
                      fontSize: 13,
                      fontWeight: FontWeight.w600,
                      color: AppTheme.textSecondary,
                    ),
                  ),
                  const SizedBox(height: 8),
                  Text(
                    _result!.formatted,
                    style: GoogleFonts.plusJakartaSans(
                      fontSize: 34,
                      fontWeight: FontWeight.w800,
                      color: AppTheme.primary,
                      letterSpacing: -0.5,
                    ),
                  ),
                  if (_result!.steps.isNotEmpty) ...[
                    const SizedBox(height: 16),
                    const Divider(color: Color(0xFFBFDBFE)),
                    const SizedBox(height: 10),
                    Align(
                      alignment: Alignment.centerLeft,
                      child: Text(
                        'Step-by-Step Breakdown:',
                        style: GoogleFonts.plusJakartaSans(
                          fontSize: 12,
                          fontWeight: FontWeight.w700,
                          color: AppTheme.textPrimary,
                        ),
                      ),
                    ),
                    const SizedBox(height: 6),
                    ..._result!.steps.map(
                      (s) => Padding(
                        padding: const EdgeInsets.symmetric(vertical: 2.5),
                        child: Row(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            const Text(
                              '• ',
                              style: TextStyle(
                                color: AppTheme.primary,
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                            Expanded(
                              child: Text(
                                s,
                                style: GoogleFonts.plusJakartaSans(
                                  fontSize: 12,
                                  color: AppTheme.textSecondary,
                                ),
                              ),
                            ),
                          ],
                        ),
                      ),
                    ),
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
