import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:google_fonts/google_fonts.dart';

import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/bouncy_button.dart';
import 'nifty50_screen.dart';
import 'technical_chart_screen.dart';

class MarketsHubScreen extends StatefulWidget {
  const MarketsHubScreen({super.key});

  @override
  State<MarketsHubScreen> createState() => _MarketsHubScreenState();
}

class _MarketsHubScreenState extends State<MarketsHubScreen> {
  int _activeTab = 0; // 0 = NIFTY 50, 1 = Technical Charts
  String _selectedSymbol = 'HINDUNILVR';

  void _switchToChart(String symbol) {
    HapticFeedback.lightImpact();
    setState(() {
      _selectedSymbol = symbol;
      _activeTab = 1;
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppTheme.scaffoldBg,
      body: SafeArea(
        child: Column(
          children: [
            // ── Top Pill Switcher: NIFTY 50 | Technical Charts ─────
            Container(
              padding: const EdgeInsets.fromLTRB(16, 12, 16, 10),
              decoration: BoxDecoration(
                color: Colors.white,
                border: Border(
                  bottom: BorderSide(
                    color: const Color(0xFFE2E8F0),
                    width: 1.0,
                  ),
                ),
              ),
              child: Row(
                children: [
                  Expanded(
                    child: Container(
                      padding: const EdgeInsets.all(4),
                      decoration: BoxDecoration(
                        color: const Color(0xFFF1F5F9),
                        borderRadius: BorderRadius.circular(14),
                      ),
                      child: Row(
                        children: [
                          Expanded(
                            child: _buildTabPill(
                              index: 0,
                              label: 'NIFTY 50',
                              icon: Icons.trending_up_rounded,
                            ),
                          ),
                          const SizedBox(width: 4),
                          Expanded(
                            child: _buildTabPill(
                              index: 1,
                              label: 'Technical Charts',
                              icon: Icons.candlestick_chart_rounded,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                ],
              ),
            ),

            // ── Content Area ──────────────────────────────────────
            Expanded(
              child: IndexedStack(
                index: _activeTab,
                children: [
                  Nifty50Screen(
                    onOpenChart: _switchToChart,
                  ),
                  TechnicalChartScreen(
                    key: ValueKey(_selectedSymbol),
                    initialSymbol: _selectedSymbol,
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildTabPill({
    required int index,
    required String label,
    required IconData icon,
  }) {
    final isSelected = _activeTab == index;

    return BouncyButton(
      onTap: () {
        HapticFeedback.lightImpact();
        setState(() => _activeTab = index);
      },
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 200),
        padding: const EdgeInsets.symmetric(vertical: 8),
        decoration: BoxDecoration(
          color: isSelected ? Colors.white : Colors.transparent,
          borderRadius: BorderRadius.circular(10),
          boxShadow: isSelected
              ? [
                  BoxShadow(
                    color: Colors.black.withValues(alpha: 0.08),
                    blurRadius: 6,
                    offset: const Offset(0, 2),
                  ),
                ]
              : null,
        ),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              icon,
              size: 16,
              color: isSelected ? const Color(0xFF1E40AF) : AppTheme.textMuted,
            ),
            const SizedBox(width: 6),
            Text(
              label,
              style: GoogleFonts.plusJakartaSans(
                fontSize: 12,
                fontWeight: isSelected ? FontWeight.w800 : FontWeight.w600,
                color: isSelected ? const Color(0xFF1E40AF) : AppTheme.textSecondary,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
