import 'package:flutter/material.dart';
import '../../models/chart_models.dart';
import 'technical_chart_painter.dart';

class InteractiveTechnicalChart extends StatefulWidget {
  final List<ChartCandleModel> candles;
  final Map<String, ChartIndicatorModel> indicatorMap;
  final bool isDark;
  final bool isLineChart;
  final bool showEma5;
  final bool showEma13;
  final bool showEma26;
  final bool showSma200;
  final bool showVolume;
  final bool showRsi;
  final String period; // '6m', '1y', 'all'
  final void Function(ChartCandleModel candle, ChartIndicatorModel? indicator)? onInspect;

  const InteractiveTechnicalChart({
    super.key,
    required this.candles,
    required this.indicatorMap,
    required this.isDark,
    required this.isLineChart,
    required this.showEma5,
    required this.showEma13,
    required this.showEma26,
    required this.showSma200,
    required this.showVolume,
    required this.showRsi,
    required this.period,
    this.onInspect,
  });

  @override
  State<InteractiveTechnicalChart> createState() => _InteractiveTechnicalChartState();
}

class _InteractiveTechnicalChartState extends State<InteractiveTechnicalChart> {
  int _visibleCount = 80;
  int _startIndex = 0;
  Offset? _crosshairPosition;

  int _baseVisibleCount = 80;

  @override
  void initState() {
    super.initState();
    _applyPeriod();
  }

  @override
  void didUpdateWidget(covariant InteractiveTechnicalChart oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.period != widget.period || oldWidget.candles != widget.candles) {
      _applyPeriod();
    }
  }

  void _applyPeriod() {
    if (widget.candles.isEmpty) return;
    final total = widget.candles.length;

    int targetCount = total;
    if (widget.period == '6m') {
      targetCount = (total < 126) ? total : 126;
    } else if (widget.period == '1y') {
      targetCount = (total < 252) ? total : 252;
    }

    final minVisible = (total < 15) ? total : 15;
    setState(() {
      _visibleCount = targetCount.clamp(minVisible, total);
      _startIndex = (total - _visibleCount).clamp(0, total);
      _crosshairPosition = null;
    });

    if (widget.candles.isNotEmpty) {
      final lastCandle = widget.candles.last;
      final dtStr = lastCandle.date.toIso8601String().substring(0, 10);
      widget.onInspect?.call(lastCandle, widget.indicatorMap[dtStr]);
    }
  }

  @override
  Widget build(BuildContext context) {
    if (widget.candles.isEmpty) {
      return Container(
        alignment: Alignment.center,
        color: widget.isDark ? const Color(0xFF0F172A) : Colors.white,
        child: Text(
          'No candle data available',
          style: TextStyle(color: widget.isDark ? Colors.white60 : Colors.black45),
        ),
      );
    }

    return GestureDetector(
      onScaleStart: (details) {
        _baseVisibleCount = _visibleCount;
      },
      onScaleUpdate: (details) {
        if (details.pointerCount == 2) {
          // Pinch to Zoom
          final newCount = (_baseVisibleCount / details.scale).round().clamp(15, widget.candles.length);
          if (newCount != _visibleCount) {
            final diff = newCount - _visibleCount;
            setState(() {
              _visibleCount = newCount;
              _startIndex = (_startIndex - (diff ~/ 2)).clamp(0, widget.candles.length - _visibleCount);
            });
          }
        } else if (details.pointerCount == 1) {
          // Horizontal Pan / Drag
          final dx = details.focalPointDelta.dx;
          if (dx.abs() > 0.5) {
            final chartWidth = context.size?.width ?? 300;
            final candleStep = chartWidth / _visibleCount;
            final shift = (-dx / candleStep).round();

            if (shift != 0) {
              setState(() {
                _startIndex = (_startIndex + shift).clamp(0, widget.candles.length - _visibleCount);
              });
            }
          }
        }
      },
      onLongPressStart: (details) {
        setState(() {
          _crosshairPosition = details.localPosition;
        });
      },
      onLongPressMoveUpdate: (details) {
        setState(() {
          _crosshairPosition = details.localPosition;
        });
      },
      onLongPressEnd: (_) {
        // Keep crosshair active until user taps elsewhere
      },
      onTapUp: (details) {
        // Tap to position crosshair or tap again to dismiss
        if (_crosshairPosition != null) {
          setState(() => _crosshairPosition = null);
          if (widget.candles.isNotEmpty) {
            final lastCandle = widget.candles.last;
            final dtStr = lastCandle.date.toIso8601String().substring(0, 10);
            widget.onInspect?.call(lastCandle, widget.indicatorMap[dtStr]);
          }
        } else {
          setState(() {
            _crosshairPosition = details.localPosition;
          });
        }
      },
      onDoubleTap: () {
        _applyPeriod();
      },
      child: ClipRect(
        child: CustomPaint(
          size: Size.infinite,
          painter: TechnicalChartPainter(
            candles: widget.candles,
            indicatorMap: widget.indicatorMap,
            startIndex: _startIndex,
            visibleCount: _visibleCount,
            isDark: widget.isDark,
            isLineChart: widget.isLineChart,
            showEma5: widget.showEma5,
            showEma13: widget.showEma13,
            showEma26: widget.showEma26,
            showSma200: widget.showSma200,
            showVolume: widget.showVolume,
            showRsi: widget.showRsi,
            crosshairPosition: _crosshairPosition,
            onCrosshairUpdated: (c, ind) {
              widget.onInspect?.call(c, ind);
            },
          ),
        ),
      ),
    );
  }
}
