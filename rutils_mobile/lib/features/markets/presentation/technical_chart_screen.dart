import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:intl/intl.dart';

import '../../../core/widgets/bouncy_button.dart';
import '../models/chart_models.dart';
import '../services/chart_service.dart';
import 'widgets/interactive_technical_chart.dart';

class TechnicalChartScreen extends StatefulWidget {
  final String? initialSymbol;

  const TechnicalChartScreen({
    super.key,
    this.initialSymbol,
  });

  @override
  State<TechnicalChartScreen> createState() => _TechnicalChartScreenState();
}

class _TechnicalChartScreenState extends State<TechnicalChartScreen> {
  final ChartService _service = ChartService();

  List<ChartCompanyModel> _companies = [];
  String _currentSymbol = 'HINDUNILVR';

  ChartDataResponse? _chartData;
  Map<String, ChartIndicatorModel> _indicatorMap = {};

  bool _isLoading = true;
  String? _errorMessage;

  // Configuration options matching web version
  bool _isDark = false;
  bool _isLineChart = false;
  String _period = '6m'; // '6m', '1y', 'all'

  // Indicator toggles
  bool _showEma5 = true;
  bool _showEma13 = true;
  bool _showEma26 = true;
  bool _showSma200 = true;
  bool _showVolume = true;
  bool _showRsi = true;

  // Inspected candle stats
  ChartCandleModel? _inspectedCandle;
  ChartIndicatorModel? _inspectedIndicator;
  double? _inspectedChangePct;

  @override
  void initState() {
    super.initState();
    if (widget.initialSymbol != null && widget.initialSymbol!.isNotEmpty) {
      _currentSymbol = widget.initialSymbol!.toUpperCase();
    }
    _initData();
  }

  Future<void> _initData() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      final list = await _service.fetchCompanies();
      if (mounted) {
        setState(() {
          _companies = list;
          if (_companies.isNotEmpty &&
              !_companies.any((c) => c.symbol == _currentSymbol)) {
            _currentSymbol = _companies.first.symbol;
          }
        });
      }
      await _loadChartData(_currentSymbol);
    } catch (e) {
      if (mounted) {
        setState(() {
          _isLoading = false;
          _errorMessage = e.toString();
        });
      }
    }
  }

  Future<void> _loadChartData(String symbol) async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      final data = await _service.fetchChartData(symbol);
      final map = <String, ChartIndicatorModel>{};
      for (final ind in data.indicators) {
        final key = DateFormat('yyyy-MM-dd').format(ind.date);
        map[key] = ind;
      }

      if (mounted) {
        setState(() {
          _isLoading = false;
          _currentSymbol = symbol;
          _chartData = data;
          _indicatorMap = map;

          if (data.candles.isNotEmpty) {
            _inspectedCandle = data.candles.last;
            final key = DateFormat('yyyy-MM-dd').format(_inspectedCandle!.date);
            _inspectedIndicator = map[key];
            if (data.candles.length > 1) {
              final prev = data.candles[data.candles.length - 2];
              if (prev.close > 0) {
                _inspectedChangePct = ((_inspectedCandle!.close - prev.close) / prev.close) * 100;
              }
            }
          }
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _isLoading = false;
          _errorMessage = e.toString();
        });
      }
    }
  }

  void _onCandleInspected(ChartCandleModel candle, ChartIndicatorModel? indicator) {
    if (_chartData == null || _chartData!.candles.isEmpty) return;

    double? pct;
    final idx = _chartData!.candles.indexOf(candle);
    if (idx > 0) {
      final prev = _chartData!.candles[idx - 1];
      if (prev.close > 0) {
        pct = ((candle.close - prev.close) / prev.close) * 100;
      }
    }

    setState(() {
      _inspectedCandle = candle;
      _inspectedIndicator = indicator;
      _inspectedChangePct = pct;
    });
  }

  void _openCompanySearchModal() {
    HapticFeedback.lightImpact();
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (ctx) => _CompanySearchSheet(
        companies: _companies,
        currentSymbol: _currentSymbol,
        isDark: _isDark,
        onSelected: (sym) {
          Navigator.pop(ctx);
          _loadChartData(sym);
        },
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final themeBg = _isDark ? const Color(0xFF0F172A) : const Color(0xFFF8FAFC);
    final cardBg = _isDark ? const Color(0xFF1E293B) : Colors.white;
    final textColor = _isDark ? const Color(0xFFF8FAFC) : const Color(0xFF0F172A);
    final subTextColor = _isDark ? const Color(0xFF94A3B8) : const Color(0xFF64748B);

    return Scaffold(
      backgroundColor: themeBg,
      body: SafeArea(
        child: Column(
          children: [
            // ── 1. Top Strip: INK CHART Controls ─────────────────────
            _buildTopControlBar(cardBg, textColor, subTextColor),

            // ── 2. Indicators & Overlays Checkboxes ──────────────────
            _buildIndicatorsBar(cardBg, textColor),

            // ── 3. Live Crosshair HUD Stats Bar ──────────────────────
            _buildHudStatsBar(cardBg, textColor, subTextColor),

            // ── 4. Main Chart Canvas Viewport ────────────────────────
            Expanded(
              child: _isLoading
                  ? Center(
                      child: Column(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          CircularProgressIndicator(
                            color: _isDark ? const Color(0xFF38BDF8) : const Color(0xFF2C7A7B),
                          ),
                          const SizedBox(height: 12),
                          Text(
                            'Loading Technical Data & Indicators...',
                            style: GoogleFonts.plusJakartaSans(
                              color: subTextColor,
                              fontSize: 12,
                              fontWeight: FontWeight.w500,
                            ),
                          ),
                        ],
                      ),
                    )
                  : _errorMessage != null
                      ? Center(
                          child: Padding(
                            padding: const EdgeInsets.all(24),
                            child: Column(
                              mainAxisSize: MainAxisSize.min,
                              children: [
                                const Icon(Icons.error_outline_rounded, color: Colors.redAccent, size: 40),
                                const SizedBox(height: 8),
                                Text(
                                  _errorMessage!,
                                  textAlign: TextAlign.center,
                                  style: GoogleFonts.plusJakartaSans(
                                    color: Colors.redAccent,
                                    fontSize: 13,
                                  ),
                                ),
                                const SizedBox(height: 14),
                                ElevatedButton.icon(
                                  onPressed: () => _loadChartData(_currentSymbol),
                                  icon: const Icon(Icons.refresh_rounded, size: 16),
                                  label: const Text('Retry'),
                                  style: ElevatedButton.styleFrom(
                                    backgroundColor: const Color(0xFF2C7A7B),
                                    foregroundColor: Colors.white,
                                  ),
                                ),
                              ],
                            ),
                          ),
                        )
                      : InteractiveTechnicalChart(
                          candles: _chartData?.candles ?? [],
                          indicatorMap: _indicatorMap,
                          isDark: _isDark,
                          isLineChart: _isLineChart,
                          showEma5: _showEma5,
                          showEma13: _showEma13,
                          showEma26: _showEma26,
                          showSma200: _showSma200,
                          showVolume: _showVolume,
                          showRsi: _showRsi,
                          period: _period,
                          onInspect: _onCandleInspected,
                        ),
            ),

            // ── 5. Trade Advice & Signal Banner (if available) ──────
            if (_chartData != null && _chartData!.advice.isNotEmpty)
              _buildTradeAdviceBanner(_chartData!.advice.last, cardBg, textColor, subTextColor),
          ],
        ),
      ),
    );
  }

  Widget _buildTopControlBar(Color cardBg, Color textColor, Color subTextColor) {
    final topBarBg = _isDark ? const Color(0xFF1E293B) : const Color(0xFF2C7A7B);

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
      decoration: BoxDecoration(
        color: topBarBg,
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.15),
            blurRadius: 4,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Column(
        children: [
          // Row 1: INK CHART badge + Stock Selector + Update + Theme
          Row(
            children: [
              // Badge
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                decoration: BoxDecoration(
                  color: _isDark ? const Color(0xFF0F172A) : const Color(0xFF234E52),
                  borderRadius: BorderRadius.circular(6),
                  border: Border.all(
                    color: _isDark ? const Color(0xFF334155) : const Color(0xFF319795),
                  ),
                ),
                child: Text(
                  'INK CHART',
                  style: GoogleFonts.plusJakartaSans(
                    fontSize: 10,
                    fontWeight: FontWeight.w800,
                    color: const Color(0xFF81E6D9),
                    letterSpacing: 0.5,
                  ),
                ),
              ),
              const SizedBox(width: 8),

              // Stock Selector Chip (Opens Search modal)
              Expanded(
                child: BouncyButton(
                  onTap: _openCompanySearchModal,
                  child: Container(
                    padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
                    decoration: BoxDecoration(
                      color: _isDark ? const Color(0xFF0F172A) : Colors.white,
                      borderRadius: BorderRadius.circular(6),
                      border: Border.all(
                        color: _isDark ? const Color(0xFF475569) : const Color(0xFFCBD5E1),
                      ),
                    ),
                    child: Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Expanded(
                          child: Text(
                            _currentSymbol,
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                            style: GoogleFonts.plusJakartaSans(
                              fontSize: 12,
                              fontWeight: FontWeight.w800,
                              color: _isDark ? Colors.white : const Color(0xFF0F172A),
                            ),
                          ),
                        ),
                        Icon(
                          Icons.search_rounded,
                          size: 16,
                          color: _isDark ? Colors.white70 : const Color(0xFF2C7A7B),
                        ),
                      ],
                    ),
                  ),
                ),
              ),
              const SizedBox(width: 8),

              // Theme Switcher Button
              BouncyButton(
                onTap: () {
                  HapticFeedback.lightImpact();
                  setState(() => _isDark = !_isDark);
                },
                child: Container(
                  padding: const EdgeInsets.all(6),
                  decoration: BoxDecoration(
                    color: _isDark ? const Color(0xFF334155) : const Color(0xFF234E52),
                    borderRadius: BorderRadius.circular(6),
                  ),
                  child: Icon(
                    _isDark ? Icons.light_mode_rounded : Icons.dark_mode_rounded,
                    size: 16,
                    color: Colors.white,
                  ),
                ),
              ),
              const SizedBox(width: 6),

              // Refresh / Update Button
              BouncyButton(
                onTap: () => _loadChartData(_currentSymbol),
                child: Container(
                  padding: const EdgeInsets.all(6),
                  decoration: BoxDecoration(
                    color: const Color(0xFF319795),
                    borderRadius: BorderRadius.circular(6),
                  ),
                  child: const Icon(
                    Icons.refresh_rounded,
                    size: 16,
                    color: Colors.white,
                  ),
                ),
              ),
            ],
          ),

          const SizedBox(height: 6),

          // Row 2: Period (6M, 1Y, ALL) + Chart Type (Candle, Line)
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              // Period Selector
              Row(
                children: [
                  Text(
                    'Period: ',
                    style: GoogleFonts.plusJakartaSans(
                      color: Colors.white70,
                      fontSize: 11,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                  _buildPeriodPill('6M', '6m'),
                  const SizedBox(width: 4),
                  _buildPeriodPill('1Y', '1y'),
                  const SizedBox(width: 4),
                  _buildPeriodPill('ALL', 'all'),
                ],
              ),

              // Chart Type Toggle
              Row(
                children: [
                  Text(
                    'Type: ',
                    style: GoogleFonts.plusJakartaSans(
                      color: Colors.white70,
                      fontSize: 11,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                  _buildTypePill(Icons.candlestick_chart_rounded, 'Candle', !_isLineChart),
                  const SizedBox(width: 4),
                  _buildTypePill(Icons.show_chart_rounded, 'Line', _isLineChart),
                ],
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildPeriodPill(String label, String value) {
    final isSelected = _period == value;
    return BouncyButton(
      onTap: () {
        HapticFeedback.lightImpact();
        setState(() => _period = value);
      },
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 7, vertical: 3),
        decoration: BoxDecoration(
          color: isSelected
              ? const Color(0xFF319795)
              : Colors.black.withValues(alpha: 0.25),
          borderRadius: BorderRadius.circular(4),
          border: Border.all(
            color: isSelected ? const Color(0xFF81E6D9) : Colors.transparent,
            width: 0.8,
          ),
        ),
        child: Text(
          label,
          style: GoogleFonts.plusJakartaSans(
            fontSize: 10,
            fontWeight: FontWeight.w700,
            color: isSelected ? Colors.white : Colors.white70,
          ),
        ),
      ),
    );
  }

  Widget _buildTypePill(IconData icon, String label, bool isSelected) {
    return BouncyButton(
      onTap: () {
        HapticFeedback.lightImpact();
        setState(() => _isLineChart = (label == 'Line'));
      },
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 3),
        decoration: BoxDecoration(
          color: isSelected
              ? const Color(0xFF319795)
              : Colors.black.withValues(alpha: 0.25),
          borderRadius: BorderRadius.circular(4),
          border: Border.all(
            color: isSelected ? const Color(0xFF81E6D9) : Colors.transparent,
            width: 0.8,
          ),
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(icon, size: 12, color: isSelected ? Colors.white : Colors.white70),
            const SizedBox(width: 2),
            Text(
              label,
              style: GoogleFonts.plusJakartaSans(
                fontSize: 10,
                fontWeight: FontWeight.w700,
                color: isSelected ? Colors.white : Colors.white70,
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildIndicatorsBar(Color cardBg, Color textColor) {
    final barBg = _isDark ? const Color(0xFF0F172A) : const Color(0xFFF8FAFC);
    final borderColor = _isDark ? const Color(0xFF334155) : const Color(0xFFE2E8F0);

    return Container(
      width: double.infinity,
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
      decoration: BoxDecoration(
        color: barBg,
        border: Border(bottom: BorderSide(color: borderColor)),
      ),
      child: SingleChildScrollView(
        scrollDirection: Axis.horizontal,
        physics: const BouncingScrollPhysics(),
        child: Row(
          children: [
            _buildIndicatorChip('EMA 5', const Color(0xFF9F7AEA), _showEma5, () {
              setState(() => _showEma5 = !_showEma5);
            }),
            _buildIndicatorChip('EMA 13', const Color(0xFF38A169), _showEma13, () {
              setState(() => _showEma13 = !_showEma13);
            }),
            _buildIndicatorChip('EMA 26', _isDark ? const Color(0xFFCBD5E1) : const Color(0xFF1E293B), _showEma26, () {
              setState(() => _showEma26 = !_showEma26);
            }),
            _buildIndicatorChip('SMA 200', const Color(0xFFE53E3E), _showSma200, () {
              setState(() => _showSma200 = !_showSma200);
            }),
            _buildIndicatorChip('Volume', const Color(0xFF0F766E), _showVolume, () {
              setState(() => _showVolume = !_showVolume);
            }),
            _buildIndicatorChip('RSI 14', const Color(0xFF2563EB), _showRsi, () {
              setState(() => _showRsi = !_showRsi);
            }),
          ],
        ),
      ),
    );
  }

  Widget _buildIndicatorChip(String label, Color color, bool isEnabled, VoidCallback onToggle) {
    return Padding(
      padding: const EdgeInsets.only(right: 6),
      child: BouncyButton(
        onTap: () {
          HapticFeedback.lightImpact();
          onToggle();
        },
        child: Container(
          padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
          decoration: BoxDecoration(
            color: isEnabled
                ? color.withValues(alpha: _isDark ? 0.25 : 0.12)
                : (_isDark ? const Color(0xFF1E293B) : const Color(0xFFF1F5F9)),
            borderRadius: BorderRadius.circular(12),
            border: Border.all(
              color: isEnabled ? color : (_isDark ? const Color(0xFF334155) : const Color(0xFFCBD5E1)),
              width: 0.8,
            ),
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              Container(
                width: 6,
                height: 6,
                decoration: BoxDecoration(
                  shape: BoxShape.circle,
                  color: isEnabled ? color : Colors.grey,
                ),
              ),
              const SizedBox(width: 5),
              Text(
                label,
                style: GoogleFonts.plusJakartaSans(
                  fontSize: 10,
                  fontWeight: FontWeight.w700,
                  color: isEnabled ? color : (_isDark ? Colors.white54 : Colors.black45),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildHudStatsBar(Color cardBg, Color textColor, Color subTextColor) {
    final hudBg = _isDark ? const Color(0xFF1E293B) : const Color(0xFFF1F5F9);
    final borderColor = _isDark ? const Color(0xFF334155) : const Color(0xFFE2E8F0);

    final c = _inspectedCandle;
    final ind = _inspectedIndicator;
    final pct = _inspectedChangePct ?? 0.0;
    final isPos = pct >= 0;

    return Container(
      width: double.infinity,
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
      decoration: BoxDecoration(
        color: hudBg,
        border: Border(bottom: BorderSide(color: borderColor)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Row 1: Ticker Badge, Close, Change %, Date
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Row(
                children: [
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                    decoration: BoxDecoration(
                      color: _isDark ? const Color(0xFF134E4A) : const Color(0xFFCCFBF1),
                      borderRadius: BorderRadius.circular(4),
                      border: Border.all(
                        color: _isDark ? const Color(0xFF0F766E) : const Color(0xFF5EEAD4),
                      ),
                    ),
                    child: Text(
                      _currentSymbol,
                      style: GoogleFonts.plusJakartaSans(
                        fontSize: 10,
                        fontWeight: FontWeight.w800,
                        color: _isDark ? const Color(0xFF5EEAD4) : const Color(0xFF0F766E),
                      ),
                    ),
                  ),
                  const SizedBox(width: 8),
                  Text(
                    c != null ? '₹${c.close.toStringAsFixed(2)}' : '-',
                    style: GoogleFonts.jetBrainsMono(
                      fontSize: 12,
                      fontWeight: FontWeight.w800,
                      color: textColor,
                    ),
                  ),
                  const SizedBox(width: 6),
                  Text(
                    '${isPos ? '+' : ''}${pct.toStringAsFixed(2)}%',
                    style: GoogleFonts.jetBrainsMono(
                      fontSize: 11,
                      fontWeight: FontWeight.w700,
                      color: isPos ? const Color(0xFF10B981) : const Color(0xFFEF5350),
                    ),
                  ),
                ],
              ),
              Text(
                c != null ? DateFormat('dd MMM yyyy').format(c.date) : '-',
                style: GoogleFonts.jetBrainsMono(
                  fontSize: 10,
                  color: subTextColor,
                  fontWeight: FontWeight.w600,
                ),
              ),
            ],
          ),

          const SizedBox(height: 4),

          // Row 2: OHLCV values in clean monospace
          SingleChildScrollView(
            scrollDirection: Axis.horizontal,
            physics: const BouncingScrollPhysics(),
            child: Row(
              children: [
                _buildStatMetric('O', c?.open.toStringAsFixed(2) ?? '-', textColor, subTextColor),
                _buildStatMetric('H', c?.high.toStringAsFixed(2) ?? '-', textColor, subTextColor),
                _buildStatMetric('L', c?.low.toStringAsFixed(2) ?? '-', textColor, subTextColor),
                _buildStatMetric('C', c?.close.toStringAsFixed(2) ?? '-', textColor, subTextColor),
                _buildStatMetric('Vol', _formatVolume(c?.volume ?? 0), textColor, subTextColor),
                if (_showEma5 && ind != null && ind.ema5 > 0)
                  _buildStatMetric('EMA5', ind.ema5.toStringAsFixed(2), const Color(0xFF9F7AEA), subTextColor),
                if (_showEma13 && ind != null && ind.ema13 > 0)
                  _buildStatMetric('EMA13', ind.ema13.toStringAsFixed(2), const Color(0xFF38A169), subTextColor),
                if (_showEma26 && ind != null && ind.ema26 > 0)
                  _buildStatMetric('EMA26', ind.ema26.toStringAsFixed(2), textColor, subTextColor),
                if (_showSma200 && ind != null && ind.sma200 > 0)
                  _buildStatMetric('SMA200', ind.sma200.toStringAsFixed(2), const Color(0xFFE53E3E), subTextColor),
                if (_showRsi && ind != null && ind.rsi >= 0)
                  _buildStatMetric('RSI', ind.rsi.toStringAsFixed(2), const Color(0xFF2563EB), subTextColor),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildStatMetric(String label, String value, Color valColor, Color labelColor) {
    return Padding(
      padding: const EdgeInsets.only(right: 8),
      child: Text.rich(
        TextSpan(
          children: [
            TextSpan(text: '$label: ', style: TextStyle(color: labelColor, fontSize: 10, fontFamily: 'monospace')),
            TextSpan(text: value, style: TextStyle(color: valColor, fontSize: 10, fontWeight: FontWeight.bold, fontFamily: 'monospace')),
          ],
        ),
      ),
    );
  }

  String _formatVolume(double v) {
    if (v >= 1e7) return '${(v / 1e7).toStringAsFixed(2)}Cr';
    if (v >= 1e5) return '${(v / 1e5).toStringAsFixed(2)}L';
    if (v >= 1e3) return '${(v / 1e3).toStringAsFixed(1)}K';
    return v.toStringAsFixed(0);
  }

  Widget _buildTradeAdviceBanner(ChartAdviceModel advice, Color cardBg, Color textColor, Color subTextColor) {
    final isBuy = advice.advice == 'BUY';
    final badgeColor = isBuy ? const Color(0xFF10B981) : const Color(0xFFEF5350);

    return Container(
      width: double.infinity,
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
      decoration: BoxDecoration(
        color: _isDark ? const Color(0xFF1E293B) : Colors.white,
        border: Border(
          top: BorderSide(
            color: _isDark ? const Color(0xFF334155) : const Color(0xFFE2E8F0),
          ),
        ),
      ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Row(
            children: [
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                decoration: BoxDecoration(
                  color: badgeColor,
                  borderRadius: BorderRadius.circular(6),
                ),
                child: Text(
                  advice.advice,
                  style: GoogleFonts.plusJakartaSans(
                    color: Colors.white,
                    fontSize: 10,
                    fontWeight: FontWeight.w800,
                  ),
                ),
              ),
              const SizedBox(width: 10),
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text(
                    'Quantitative Strategy Signal',
                    style: GoogleFonts.plusJakartaSans(
                      color: textColor,
                      fontSize: 11,
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                  Text(
                    'Target: ₹${advice.targetPrice.toStringAsFixed(2)} | 200-SMA: ₹${advice.sma200Support.toStringAsFixed(2)}',
                    style: GoogleFonts.jetBrainsMono(
                      color: subTextColor,
                      fontSize: 10,
                    ),
                  ),
                ],
              ),
            ],
          ),
          Text(
            'SmartAPI',
            style: GoogleFonts.plusJakartaSans(
              color: subTextColor.withValues(alpha: 0.6),
              fontSize: 9,
              fontWeight: FontWeight.w600,
            ),
          ),
        ],
      ),
    );
  }
}

// ─────────────────────────────────────────────────────────────
// 101-Company Search Modal Bottom Sheet
// ─────────────────────────────────────────────────────────────
class _CompanySearchSheet extends StatefulWidget {
  final List<ChartCompanyModel> companies;
  final String currentSymbol;
  final bool isDark;
  final ValueChanged<String> onSelected;

  const _CompanySearchSheet({
    required this.companies,
    required this.currentSymbol,
    required this.isDark,
    required this.onSelected,
  });

  @override
  State<_CompanySearchSheet> createState() => _CompanySearchSheetState();
}

class _CompanySearchSheetState extends State<_CompanySearchSheet> {
  final TextEditingController _searchController = TextEditingController();
  List<ChartCompanyModel> _filtered = [];

  @override
  void initState() {
    super.initState();
    _filtered = widget.companies;
    _searchController.addListener(_filter);
  }

  void _filter() {
    final query = _searchController.text.trim().toUpperCase();
    setState(() {
      if (query.isEmpty) {
        _filtered = widget.companies;
      } else {
        _filtered = widget.companies.where((c) {
          final sym = c.symbol.toUpperCase();
          final name = c.displayName.toUpperCase();
          final ind = (c.industry ?? '').toUpperCase();
          return sym.contains(query) || name.contains(query) || ind.contains(query);
        }).toList();
      }
    });
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final sheetBg = widget.isDark ? const Color(0xFF1E293B) : Colors.white;
    final itemBg = widget.isDark ? const Color(0xFF0F172A) : const Color(0xFFF8FAFC);
    final textColor = widget.isDark ? Colors.white : const Color(0xFF0F172A);
    final subColor = widget.isDark ? const Color(0xFF94A3B8) : const Color(0xFF64748B);

    return Container(
      height: MediaQuery.of(context).size.height * 0.75,
      decoration: BoxDecoration(
        color: sheetBg,
        borderRadius: const BorderRadius.vertical(top: Radius.circular(24)),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.3),
            blurRadius: 20,
            offset: const Offset(0, -4),
          ),
        ],
      ),
      child: Column(
        children: [
          // Drag handle
          Container(
            margin: const EdgeInsets.only(top: 10, bottom: 6),
            width: 40,
            height: 4,
            decoration: BoxDecoration(
              color: Colors.grey.withValues(alpha: 0.4),
              borderRadius: BorderRadius.circular(2),
            ),
          ),

          // Header
          Padding(
            padding: const EdgeInsets.fromLTRB(18, 6, 18, 12),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  'Select Stock (${widget.companies.length} Available)',
                  style: GoogleFonts.plusJakartaSans(
                    fontSize: 16,
                    fontWeight: FontWeight.w800,
                    color: textColor,
                  ),
                ),
                IconButton(
                  onPressed: () => Navigator.pop(context),
                  icon: const Icon(Icons.close_rounded, size: 20),
                  color: subColor,
                ),
              ],
            ),
          ),

          // Search Field
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
            child: TextField(
              controller: _searchController,
              autofocus: true,
              style: GoogleFonts.plusJakartaSans(color: textColor, fontSize: 14),
              decoration: InputDecoration(
                hintText: 'Search 101 tickers, company, sector...',
                hintStyle: GoogleFonts.plusJakartaSans(color: subColor, fontSize: 13),
                prefixIcon: Icon(Icons.search_rounded, color: subColor, size: 20),
                suffixIcon: _searchController.text.isNotEmpty
                    ? IconButton(
                        icon: const Icon(Icons.clear_rounded, size: 18),
                        onPressed: () => _searchController.clear(),
                      )
                    : null,
                filled: true,
                fillColor: itemBg,
                contentPadding: const EdgeInsets.symmetric(vertical: 10, horizontal: 14),
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(12),
                  borderSide: BorderSide(
                    color: widget.isDark ? const Color(0xFF334155) : const Color(0xFFCBD5E1),
                  ),
                ),
                enabledBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(12),
                  borderSide: BorderSide(
                    color: widget.isDark ? const Color(0xFF334155) : const Color(0xFFCBD5E1),
                  ),
                ),
                focusedBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(12),
                  borderSide: const BorderSide(color: Color(0xFF2C7A7B), width: 1.5),
                ),
              ),
            ),
          ),

          const SizedBox(height: 8),

          // Company List
          Expanded(
            child: _filtered.isEmpty
                ? Center(
                    child: Text(
                      'No companies match your search',
                      style: GoogleFonts.plusJakartaSans(color: subColor, fontSize: 13),
                    ),
                  )
                : ListView.builder(
                    itemCount: _filtered.length,
                    physics: const BouncingScrollPhysics(),
                    padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 6),
                    itemBuilder: (ctx, i) {
                      final c = _filtered[i];
                      final isSelected = c.symbol == widget.currentSymbol;

                      return Padding(
                        padding: const EdgeInsets.only(bottom: 6),
                        child: BouncyButton(
                          onTap: () {
                            HapticFeedback.selectionClick();
                            widget.onSelected(c.symbol);
                          },
                          child: Container(
                            padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
                            decoration: BoxDecoration(
                              color: isSelected
                                  ? (widget.isDark ? const Color(0xFF1E3A8A) : const Color(0xFFE0F2FE))
                                  : itemBg,
                              borderRadius: BorderRadius.circular(12),
                              border: Border.all(
                                color: isSelected
                                    ? const Color(0xFF3B82F6)
                                    : (widget.isDark ? const Color(0xFF334155) : const Color(0xFFE2E8F0)),
                              ),
                            ),
                            child: Row(
                              children: [
                                // Symbol Badge
                                Container(
                                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                                  decoration: BoxDecoration(
                                    color: widget.isDark ? const Color(0xFF134E4A) : const Color(0xFFCCFBF1),
                                    borderRadius: BorderRadius.circular(6),
                                  ),
                                  child: Text(
                                    c.symbol,
                                    style: GoogleFonts.jetBrainsMono(
                                      fontSize: 11,
                                      fontWeight: FontWeight.w800,
                                      color: widget.isDark ? const Color(0xFF5EEAD4) : const Color(0xFF0F766E),
                                    ),
                                  ),
                                ),
                                const SizedBox(width: 12),

                                // Name & Industry
                                Expanded(
                                  child: Column(
                                    crossAxisAlignment: CrossAxisAlignment.start,
                                    children: [
                                      Text(
                                        c.displayName,
                                        maxLines: 1,
                                        overflow: TextOverflow.ellipsis,
                                        style: GoogleFonts.plusJakartaSans(
                                          fontSize: 13,
                                          fontWeight: FontWeight.w700,
                                          color: textColor,
                                        ),
                                      ),
                                      if (c.industry != null && c.industry!.isNotEmpty)
                                        Text(
                                          c.industry!,
                                          style: GoogleFonts.plusJakartaSans(
                                            fontSize: 10,
                                            color: subColor,
                                          ),
                                        ),
                                    ],
                                  ),
                                ),

                                if (isSelected)
                                  const Icon(
                                    Icons.check_circle_rounded,
                                    color: Color(0xFF3B82F6),
                                    size: 18,
                                  ),
                              ],
                            ),
                          ),
                        ),
                      );
                    },
                  ),
          ),
        ],
      ),
    );
  }
}
