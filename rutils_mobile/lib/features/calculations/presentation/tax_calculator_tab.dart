import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/utils/formatters.dart';
import '../../../core/widgets/app_dropdown.dart';
import '../../../core/widgets/custom_text_field.dart';
import '../../../core/widgets/glass_card.dart';
import '../../../core/widgets/gradient_button.dart';
import '../../../core/widgets/section_header.dart';
import '../models/calculation_models.dart';
import '../services/calculation_service.dart';

class TaxCalculatorTab extends StatefulWidget {
  const TaxCalculatorTab({super.key});

  @override
  State<TaxCalculatorTab> createState() => _TaxCalculatorTabState();
}

class _TaxCalculatorTabState extends State<TaxCalculatorTab> {
  final CalculationService _service = CalculationService();

  final _amountController = TextEditingController(text: '1000');
  final _rateController = TextEditingController(text: '18');
  final _netController = TextEditingController(text: '1000');
  final _grossController = TextEditingController(text: '1180');
  final _taxPaidController = TextEditingController(text: '180');
  final _incomeController = TextEditingController(text: '800000');

  String _mode = 'add_tax'; // add_tax | remove_tax | split_gst | reverse_gst | find_rate | income_tax
  String _regime = 'new'; // new | old
  bool _isLoading = false;
  TaxResultModel? _result;
  String? _errorMessage;

  @override
  void dispose() {
    _amountController.dispose();
    _rateController.dispose();
    _netController.dispose();
    _grossController.dispose();
    _taxPaidController.dispose();
    _incomeController.dispose();
    super.dispose();
  }

  Future<void> _calculate() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      TaxResultModel res;
      if (_mode == 'add_tax' || _mode == 'remove_tax' || _mode == 'split_gst') {
        final amt = double.tryParse(_amountController.text) ?? 0;
        final rate = double.tryParse(_rateController.text) ?? 0;
        res = await _service.calculateTax(mode: _mode, amount: amt, rate: rate);
      } else if (_mode == 'reverse_gst') {
        final tp = double.tryParse(_taxPaidController.text) ?? 0;
        final rate = double.tryParse(_rateController.text) ?? 0;
        res = await _service.calculateTax(mode: _mode, taxPaid: tp, rate: rate);
      } else if (_mode == 'find_rate') {
        final net = double.tryParse(_netController.text) ?? 0;
        final gross = double.tryParse(_grossController.text) ?? 0;
        res = await _service.calculateTax(mode: _mode, net: net, gross: gross);
      } else {
        // income_tax
        final inc = double.tryParse(_incomeController.text) ?? 0;
        res = await _service.calculateTax(mode: _mode, income: inc, regime: _regime);
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
            title: 'Tax / GST Calculator',
            subtitle: 'Compute GST, VAT, Reverse Tax, and Income Tax deductions',
            icon: Icons.receipt_long_rounded,
          ),
          const SizedBox(height: 12),
          GlassCard(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                AppDropdown<String>(
                  label: 'Calculation Mode',
                  value: _mode,
                  items: const [
                    AppDropdownItem(
                      value: 'add_tax',
                      label: 'Add Tax (Net → Gross)',
                      subtitle: 'Compute tax-inclusive total from base amount',
                      icon: Icons.add_circle_outline_rounded,
                    ),
                    AppDropdownItem(
                      value: 'remove_tax',
                      label: 'Remove Tax (Gross → Net)',
                      subtitle: 'Extract base amount and tax from gross total',
                      icon: Icons.remove_circle_outline_rounded,
                    ),
                    AppDropdownItem(
                      value: 'split_gst',
                      label: 'Split GST (CGST + SGST)',
                      subtitle: 'Equal breakdown of central & state GST',
                      icon: Icons.pie_chart_rounded,
                    ),
                    AppDropdownItem(
                      value: 'reverse_gst',
                      label: 'Reverse GST from Tax Paid',
                      subtitle: 'Recover net taxable amount from tax paid',
                      icon: Icons.history_rounded,
                    ),
                    AppDropdownItem(
                      value: 'find_rate',
                      label: 'Find Effective Tax Rate',
                      subtitle: 'Calculate exact % rate from net & gross',
                      icon: Icons.percent_rounded,
                    ),
                    AppDropdownItem(
                      value: 'income_tax',
                      label: 'Income Tax (India Slabs)',
                      subtitle: 'New vs Old progressive tax slabs',
                      icon: Icons.account_balance_rounded,
                    ),
                  ],
                  modalTitle: 'Choose Tax Calculation Mode',
                  onChanged: (v) {
                    setState(() {
                      _mode = v;
                      _result = null;
                    });
                  },
                ),
                const SizedBox(height: 16),

                if (_mode == 'add_tax' || _mode == 'remove_tax' || _mode == 'split_gst') ...[
                  CustomTextField(
                    controller: _amountController,
                    label: _mode == 'remove_tax' ? 'Gross Amount (Tax-inclusive)' : 'Base Amount (Tax-exclusive)',
                    hint: 'e.g. 1000',
                    keyboardType: const TextInputType.numberWithOptions(decimal: true),
                  ),
                  const SizedBox(height: 14),
                  CustomTextField(
                    controller: _rateController,
                    label: 'Tax / GST Rate (%)',
                    hint: 'e.g. 18',
                    keyboardType: const TextInputType.numberWithOptions(decimal: true),
                  ),
                ] else if (_mode == 'reverse_gst') ...[
                  CustomTextField(
                    controller: _taxPaidController,
                    label: 'Tax Amount Paid (₹)',
                    hint: 'e.g. 180',
                    keyboardType: const TextInputType.numberWithOptions(decimal: true),
                  ),
                  const SizedBox(height: 14),
                  CustomTextField(
                    controller: _rateController,
                    label: 'GST Rate (%)',
                    hint: 'e.g. 18',
                    keyboardType: const TextInputType.numberWithOptions(decimal: true),
                  ),
                ] else if (_mode == 'find_rate') ...[
                  CustomTextField(
                    controller: _netController,
                    label: 'Net / Base Amount (₹)',
                    hint: 'e.g. 1000',
                    keyboardType: const TextInputType.numberWithOptions(decimal: true),
                  ),
                  const SizedBox(height: 14),
                  CustomTextField(
                    controller: _grossController,
                    label: 'Gross / Total Amount (₹)',
                    hint: 'e.g. 1180',
                    keyboardType: const TextInputType.numberWithOptions(decimal: true),
                  ),
                ] else if (_mode == 'income_tax') ...[
                  CustomTextField(
                    controller: _incomeController,
                    label: 'Annual Taxable Income (₹)',
                    hint: 'e.g. 800000',
                    keyboardType: const TextInputType.numberWithOptions(decimal: true),
                  ),
                  const SizedBox(height: 14),
                  Text('Tax Regime', style: GoogleFonts.plusJakartaSans(fontSize: 12, fontWeight: FontWeight.w600)),
                  const SizedBox(height: 6),
                  Row(
                    children: [
                      Expanded(
                        child: ChoiceChip(
                          label: const Text('New Regime'),
                          selected: _regime == 'new',
                          onSelected: (s) => setState(() => _regime = 'new'),
                        ),
                      ),
                      const SizedBox(width: 10),
                      Expanded(
                        child: ChoiceChip(
                          label: const Text('Old Regime'),
                          selected: _regime == 'old',
                          onSelected: (s) => setState(() => _regime = 'old'),
                        ),
                      ),
                    ],
                  ),
                ],

                const SizedBox(height: 24),
                GradientButton(
                  text: 'Calculate Tax',
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
                  Text('Total Tax Amount', style: GoogleFonts.plusJakartaSans(fontSize: 13, fontWeight: FontWeight.w600, color: AppTheme.textSecondary)),
                  const SizedBox(height: 6),
                  Text(
                    Formatters.inr(_result!.taxAmount),
                    style: GoogleFonts.plusJakartaSans(
                      fontSize: 34,
                      fontWeight: FontWeight.w800,
                      color: AppTheme.primary,
                      letterSpacing: -0.5,
                    ),
                  ),
                  const SizedBox(height: 16),
                  Row(
                    children: [
                      Expanded(
                        child: Container(
                          padding: const EdgeInsets.all(12),
                          decoration: BoxDecoration(color: Colors.white, borderRadius: BorderRadius.circular(12)),
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text('Net Amount', style: GoogleFonts.plusJakartaSans(fontSize: 11, color: AppTheme.textMuted)),
                              const SizedBox(height: 4),
                              Text(Formatters.inr(_result!.netAmount), style: GoogleFonts.plusJakartaSans(fontSize: 14, fontWeight: FontWeight.w700)),
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
                              Text('Gross Amount', style: GoogleFonts.plusJakartaSans(fontSize: 11, color: AppTheme.textMuted)),
                              const SizedBox(height: 4),
                              Text(Formatters.inr(_result!.grossAmount), style: GoogleFonts.plusJakartaSans(fontSize: 14, fontWeight: FontWeight.w700)),
                            ],
                          ),
                        ),
                      ),
                    ],
                  ),
                  if (_result!.steps.isNotEmpty) ...[
                    const SizedBox(height: 16),
                    const Divider(color: Color(0xFFBFDBFE)),
                    const SizedBox(height: 10),
                    Align(
                      alignment: Alignment.centerLeft,
                      child: Text('Calculation Steps:', style: GoogleFonts.plusJakartaSans(fontSize: 12, fontWeight: FontWeight.w700)),
                    ),
                    const SizedBox(height: 6),
                    ..._result!.steps.map(
                      (s) => Padding(
                        padding: const EdgeInsets.symmetric(vertical: 2),
                        child: Row(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            const Text('• ', style: TextStyle(color: AppTheme.primary, fontWeight: FontWeight.bold)),
                            Expanded(child: Text(s, style: GoogleFonts.plusJakartaSans(fontSize: 12, color: AppTheme.textSecondary))),
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
