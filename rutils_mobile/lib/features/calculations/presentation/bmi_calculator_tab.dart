import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/app_dropdown.dart';
import '../../../core/widgets/custom_text_field.dart';
import '../../../core/widgets/glass_card.dart';
import '../../../core/widgets/gradient_button.dart';
import '../../../core/widgets/section_header.dart';
import '../../../core/widgets/status_badge.dart';
import '../models/calculation_models.dart';
import '../services/calculation_service.dart';

class BmiCalculatorTab extends StatefulWidget {
  const BmiCalculatorTab({super.key});

  @override
  State<BmiCalculatorTab> createState() => _BmiCalculatorTabState();
}

class _BmiCalculatorTabState extends State<BmiCalculatorTab> {
  final CalculationService _service = CalculationService();
  final _heightController = TextEditingController(text: '175');
  final _weightController = TextEditingController(text: '70');

  String _heightUnit = 'centimeters'; // Backend expects: centimeters | meters | inches | feet
  String _weightUnit = 'kilograms'; // Backend expects: kilograms | pounds | grams
  bool _isLoading = false;
  BMIResult? _result;
  String? _errorMessage;

  final List<AppDropdownItem<String>> _heightUnits = const [
    AppDropdownItem(
      value: 'centimeters',
      label: 'Centimeters (cm)',
      subtitle: 'Standard metric height',
      icon: Icons.straighten_rounded,
    ),
    AppDropdownItem(
      value: 'meters',
      label: 'Meters (m)',
      subtitle: 'Metric SI unit',
      icon: Icons.height_rounded,
    ),
    AppDropdownItem(
      value: 'feet',
      label: 'Feet (ft)',
      subtitle: 'Imperial height',
      icon: Icons.straighten_rounded,
    ),
    AppDropdownItem(
      value: 'inches',
      label: 'Inches (in)',
      subtitle: 'Imperial inches',
      icon: Icons.straighten_rounded,
    ),
  ];

  final List<AppDropdownItem<String>> _weightUnits = const [
    AppDropdownItem(
      value: 'kilograms',
      label: 'Kilograms (kg)',
      subtitle: 'Standard metric weight',
      icon: Icons.fitness_center_rounded,
    ),
    AppDropdownItem(
      value: 'pounds',
      label: 'Pounds (lbs)',
      subtitle: 'Imperial weight',
      icon: Icons.scale_rounded,
    ),
    AppDropdownItem(
      value: 'grams',
      label: 'Grams (g)',
      subtitle: 'Precision metric grams',
      icon: Icons.fitness_center_rounded,
    ),
  ];

  @override
  void dispose() {
    _heightController.dispose();
    _weightController.dispose();
    super.dispose();
  }

  Color _getCategoryColor(String cat) {
    switch (cat.toLowerCase()) {
      case 'underweight':
        return const Color(0xFF0284C7);
      case 'normal weight':
        return AppTheme.success;
      case 'overweight':
        return AppTheme.warning;
      case 'obesity':
      case 'obese':
        return AppTheme.danger;
      default:
        return AppTheme.primary;
    }
  }

  Future<void> _calculate() async {
    final h = double.tryParse(_heightController.text.trim());
    final w = double.tryParse(_weightController.text.trim());

    if (h == null || h <= 0 || w == null || w <= 0) {
      setState(() => _errorMessage = 'Please enter valid positive numbers for height & weight');
      return;
    }

    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      final res = await _service.calculateBMI(
        height: h,
        heightUnit: _heightUnit,
        weight: w,
        weightUnit: _weightUnit,
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
    return SingleChildScrollView(
      physics: const BouncingScrollPhysics(),
      padding: const EdgeInsets.all(20),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          const SectionHeader(
            title: 'BMI Calculator',
            subtitle: 'Calculate Body Mass Index and assess healthy weight classification',
            icon: Icons.monitor_weight_rounded,
          ),
          const SizedBox(height: 12),
          GlassCard(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // Height Field
                CustomTextField(
                  controller: _heightController,
                  label: 'Height',
                  hint: _heightUnit == 'centimeters'
                      ? 'e.g. 175'
                      : _heightUnit == 'meters'
                          ? 'e.g. 1.75'
                          : _heightUnit == 'feet'
                              ? 'e.g. 5.75'
                              : 'e.g. 69',
                  prefixIcon: Icons.height_rounded,
                  keyboardType: const TextInputType.numberWithOptions(decimal: true),
                ),
                const SizedBox(height: 12),

                // Height Unit Dropdown
                AppDropdown<String>(
                  label: 'Height Unit',
                  value: _heightUnit,
                  items: _heightUnits,
                  modalTitle: 'Select Height Unit',
                  onChanged: (val) => setState(() => _heightUnit = val),
                ),
                const SizedBox(height: 16),

                // Weight Field
                CustomTextField(
                  controller: _weightController,
                  label: 'Weight',
                  hint: _weightUnit == 'kilograms' ? 'e.g. 70' : 'e.g. 154',
                  prefixIcon: Icons.fitness_center_rounded,
                  keyboardType: const TextInputType.numberWithOptions(decimal: true),
                ),
                const SizedBox(height: 12),

                // Weight Unit Dropdown
                AppDropdown<String>(
                  label: 'Weight Unit',
                  value: _weightUnit,
                  items: _weightUnits,
                  modalTitle: 'Select Weight Unit',
                  onChanged: (val) => setState(() => _weightUnit = val),
                ),
                const SizedBox(height: 24),

                GradientButton(
                  text: 'Calculate BMI',
                  isLoading: _isLoading,
                  icon: Icons.fitness_center_rounded,
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
              padding: const EdgeInsets.all(24),
              child: Column(
                children: [
                  Text(
                    'Your BMI Score',
                    style: GoogleFonts.plusJakartaSans(
                      fontSize: 13,
                      fontWeight: FontWeight.w600,
                      color: AppTheme.textSecondary,
                    ),
                  ),
                  const SizedBox(height: 8),
                  Text(
                    _result!.bmi.toStringAsFixed(1),
                    style: GoogleFonts.plusJakartaSans(
                      fontSize: 42,
                      fontWeight: FontWeight.w800,
                      color: _getCategoryColor(_result!.category),
                    ),
                  ),
                  const SizedBox(height: 10),
                  StatusBadge(
                    label: _result!.category,
                    color: _getCategoryColor(_result!.category),
                  ),
                  const SizedBox(height: 20),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceAround,
                    children: [
                      _buildScaleIndicator('Under', '< 18.5', const Color(0xFF0284C7)),
                      _buildScaleIndicator('Normal', '18.5 - 24.9', AppTheme.success),
                      _buildScaleIndicator('Over', '25 - 29.9', AppTheme.warning),
                      _buildScaleIndicator('Obese', '≥ 30', AppTheme.danger),
                    ],
                  ),
                ],
              ),
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildScaleIndicator(String label, String range, Color color) {
    return Column(
      children: [
        Container(
          width: 8,
          height: 8,
          decoration: BoxDecoration(color: color, shape: BoxShape.circle),
        ),
        const SizedBox(height: 4),
        Text(
          label,
          style: GoogleFonts.plusJakartaSans(
            fontSize: 11,
            fontWeight: FontWeight.w700,
            color: color,
          ),
        ),
        Text(
          range,
          style: GoogleFonts.plusJakartaSans(fontSize: 10, color: AppTheme.textMuted),
        ),
      ],
    );
  }
}
