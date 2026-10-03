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
import '../services/conversion_service.dart';

class MeasurementTab extends StatefulWidget {
  const MeasurementTab({super.key});

  @override
  State<MeasurementTab> createState() => _MeasurementTabState();
}

class _MeasurementTabState extends State<MeasurementTab> {
  final ConversionService _service = ConversionService();
  final _valueController = TextEditingController(text: '1');

  static const Map<String, List<String>> _units = {
    'length': ['meters', 'kilometers', 'centimeters', 'millimeters', 'miles', 'yards', 'feet', 'inches'],
    'weight': ['kilograms', 'grams', 'milligrams', 'pounds', 'ounces'],
    'volume': ['liters', 'milliliters', 'gallons', 'quarts', 'pints', 'fluid_ounces'],
    'area': ['square_meters', 'square_kilometers', 'hectares', 'acres', 'square_feet', 'square_miles'],
    'time': ['seconds', 'minutes', 'hours', 'days', 'weeks'],
    'temperature': ['celsius', 'fahrenheit', 'kelvin'],
    'speed': ['meters_per_second', 'kilometers_per_hour', 'miles_per_hour', 'feet_per_second', 'knots'],
    'data': ['bytes', 'kilobytes', 'megabytes', 'gigabytes', 'terabytes', 'petabytes', 'bits'],
  };

  String _category = 'length';
  late String _fromUnit;
  late String _toUnit;
  bool _isLoading = false;
  double? _result;
  String? _errorMessage;

  @override
  void initState() {
    super.initState();
    _fromUnit = _units[_category]![0];
    _toUnit = _units[_category]![1];
  }

  @override
  void dispose() {
    _valueController.dispose();
    super.dispose();
  }

  void _onCategoryChanged(String newCat) {
    setState(() {
      _category = newCat;
      _fromUnit = _units[newCat]![0];
      _toUnit = _units[newCat]!.length > 1 ? _units[newCat]![1] : _units[newCat]![0];
      _result = null;
    });
  }

  void _swapUnits() {
    setState(() {
      final temp = _fromUnit;
      _fromUnit = _toUnit;
      _toUnit = temp;
      _result = null;
    });
  }

  String _formatUnitName(String u) {
    return u.split('_').map((w) => w[0].toUpperCase() + w.substring(1)).join(' ');
  }

  Future<void> _convert() async {
    final val = double.tryParse(_valueController.text);
    if (val == null) {
      setState(() => _errorMessage = 'Please enter a valid number');
      return;
    }

    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      final res = await _service.convertMeasurement(
        category: _category,
        fromUnit: _fromUnit,
        toUnit: _toUnit,
        value: val,
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
    final currentUnits = _units[_category] ?? [];

    return SingleChildScrollView(
      physics: const BouncingScrollPhysics(),
      padding: const EdgeInsets.all(20),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          const SectionHeader(
            title: 'Unit Converter',
            subtitle: 'Convert between physical units across multiple categories',
            icon: Icons.straighten_rounded,
          ),
          const SizedBox(height: 12),

          // Categories horizontal scroll
          SizedBox(
            height: 40,
            child: ListView(
              scrollDirection: Axis.horizontal,
              physics: const BouncingScrollPhysics(),
              children: _units.keys.map((cat) {
                final isSelected = cat == _category;
                return Padding(
                  padding: const EdgeInsets.only(right: 8),
                  child: BouncyButton(
                    onTap: () => _onCategoryChanged(cat),
                    child: Container(
                      padding: const EdgeInsets.symmetric(horizontal: 16),
                      decoration: BoxDecoration(
                        color: isSelected ? AppTheme.primary : Colors.white,
                        borderRadius: BorderRadius.circular(12),
                        border: Border.all(
                          color: isSelected ? AppTheme.primary : const Color(0xFFE2E8F0),
                        ),
                        boxShadow: isSelected ? AppTheme.softShadow : null,
                      ),
                      alignment: Alignment.center,
                      child: Text(
                        _formatUnitName(cat),
                        style: GoogleFonts.plusJakartaSans(
                          fontSize: 13,
                          fontWeight: FontWeight.w600,
                          color: isSelected ? Colors.white : AppTheme.textSecondary,
                        ),
                      ),
                    ),
                  ),
                );
              }).toList(),
            ),
          ),
          const SizedBox(height: 16),

          GlassCard(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                CustomTextField(
                  controller: _valueController,
                  label: 'Value to Convert',
                  hint: 'e.g. 100',
                  keyboardType: const TextInputType.numberWithOptions(decimal: true),
                ),
                const SizedBox(height: 16),

                // From Unit
                AppDropdown<String>(
                  label: 'From Unit',
                  value: _fromUnit,
                  items: currentUnits.map((u) {
                    return AppDropdownItem(
                      value: u,
                      label: _formatUnitName(u),
                      icon: Icons.straighten_rounded,
                    );
                  }).toList(),
                  modalTitle: 'Select Source Unit',
                  onChanged: (val) => setState(() => _fromUnit = val),
                ),
                const SizedBox(height: 12),

                // Swap Button
                Center(
                  child: BouncyButton(
                    onTap: _swapUnits,
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
                const SizedBox(height: 12),

                // To Unit
                AppDropdown<String>(
                  label: 'To Unit',
                  value: _toUnit,
                  items: currentUnits.map((u) {
                    return AppDropdownItem(
                      value: u,
                      label: _formatUnitName(u),
                      icon: Icons.straighten_rounded,
                    );
                  }).toList(),
                  modalTitle: 'Select Target Unit',
                  onChanged: (val) => setState(() => _toUnit = val),
                ),
                const SizedBox(height: 24),

                GradientButton(
                  text: 'Convert',
                  isLoading: _isLoading,
                  icon: Icons.calculate_rounded,
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
              child: Text(
                _errorMessage!,
                style: GoogleFonts.plusJakartaSans(color: AppTheme.danger, fontSize: 13),
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
                    'Result',
                    style: GoogleFonts.plusJakartaSans(
                      fontSize: 13,
                      fontWeight: FontWeight.w600,
                      color: AppTheme.textSecondary,
                    ),
                  ),
                  const SizedBox(height: 6),
                  Text(
                    Formatters.number(_result!, decimals: 4),
                    style: GoogleFonts.plusJakartaSans(
                      fontSize: 32,
                      fontWeight: FontWeight.w800,
                      color: AppTheme.primary,
                      letterSpacing: -0.5,
                    ),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    '${_valueController.text} ${_formatUnitName(_fromUnit)} = ${Formatters.number(_result!, decimals: 4)} ${_formatUnitName(_toUnit)}',
                    style: GoogleFonts.plusJakartaSans(
                      fontSize: 13,
                      color: AppTheme.textMuted,
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
