import 'package:flutter/material.dart';
import 'package:intl/intl.dart' hide TextDirection;
import '../../models/chart_models.dart';

class TechnicalChartPainter extends CustomPainter {
  final List<ChartCandleModel> candles;
  final Map<String, ChartIndicatorModel> indicatorMap;
  final int startIndex;
  final int visibleCount;
  final bool isDark;
  final bool isLineChart;
  final bool showEma5;
  final bool showEma13;
  final bool showEma26;
  final bool showSma200;
  final bool showVolume;
  final bool showRsi;
  final Offset? crosshairPosition;
  final void Function(ChartCandleModel candle, ChartIndicatorModel? indicator)? onCrosshairUpdated;

  TechnicalChartPainter({
    required this.candles,
    required this.indicatorMap,
    required this.startIndex,
    required this.visibleCount,
    required this.isDark,
    required this.isLineChart,
    required this.showEma5,
    required this.showEma13,
    required this.showEma26,
    required this.showSma200,
    required this.showVolume,
    required this.showRsi,
    this.crosshairPosition,
    this.onCrosshairUpdated,
  });

  @override
  void paint(Canvas canvas, Size size) {
    if (candles.isEmpty || visibleCount <= 0) return;

    final priceScaleWidth = 54.0;
    final dateScaleHeight = 22.0;
    final chartWidth = size.width - priceScaleWidth;

    final rsiHeight = showRsi ? 80.0 : 0.0;
    final rsiGap = showRsi ? 10.0 : 0.0;
    final mainChartHeight = size.height - rsiHeight - rsiGap - dateScaleHeight;
    final rsiTop = mainChartHeight + rsiGap;

    // Palette
    final bgColor = isDark ? const Color(0xFF0F172A) : Colors.white;
    final gridColor = isDark ? const Color(0xFF1E293B) : const Color(0xFFF1F5F9);
    final borderColor = isDark ? const Color(0xFF334155) : const Color(0xFFCBD5E1);
    final textColor = isDark ? const Color(0xFF94A3B8) : const Color(0xFF64748B);

    final bullColor = const Color(0xFF26A69A);
    final bearColor = const Color(0xFFEF5350);

    final ema5Color = isDark ? const Color(0xFFC084FC) : const Color(0xFF9F7AEA);
    final ema13Color = isDark ? const Color(0xFF34D399) : const Color(0xFF38A169);
    final ema26Color = isDark ? const Color(0xFFCBD5E1) : const Color(0xFF1E293B);
    final sma200Color = isDark ? const Color(0xFFF87171) : const Color(0xFFE53E3E);
    final rsiColor = isDark ? const Color(0xFF38BDF8) : const Color(0xFF2563EB);

    // 1. Draw Background
    canvas.drawRect(
      Rect.fromLTWH(0, 0, size.width, size.height),
      Paint()..color = bgColor,
    );

    // Visible Window Calculation
    final count = visibleCount.clamp(1, candles.length);
    final start = startIndex.clamp(0, candles.length - count);
    final end = (start + count).clamp(start + 1, candles.length);
    final visibleCandles = candles.sublist(start, end);

    if (visibleCandles.isEmpty) return;

    // Determine min/max price within visible candles & visible indicators
    double minP = visibleCandles.first.low;
    double maxP = visibleCandles.first.high;
    double maxVol = 1.0;

    for (final c in visibleCandles) {
      if (c.low < minP) minP = c.low;
      if (c.high > maxP) maxP = c.high;
      if (c.volume > maxVol) maxVol = c.volume;

      final dtStr = DateFormat('yyyy-MM-dd').format(c.date);
      final ind = indicatorMap[dtStr];
      if (ind != null) {
        if (showEma5 && ind.ema5 > 0) {
          if (ind.ema5 < minP) minP = ind.ema5;
          if (ind.ema5 > maxP) maxP = ind.ema5;
        }
        if (showEma13 && ind.ema13 > 0) {
          if (ind.ema13 < minP) minP = ind.ema13;
          if (ind.ema13 > maxP) maxP = ind.ema13;
        }
        if (showEma26 && ind.ema26 > 0) {
          if (ind.ema26 < minP) minP = ind.ema26;
          if (ind.ema26 > maxP) maxP = ind.ema26;
        }
        if (showSma200 && ind.sma200 > 0) {
          if (ind.sma200 < minP) minP = ind.sma200;
          if (ind.sma200 > maxP) maxP = ind.sma200;
        }
      }
    }

    // Add 6% margin to prices for clean breathing room
    final pRange = (maxP - minP) <= 0 ? 1.0 : (maxP - minP);
    minP -= pRange * 0.04;
    maxP += pRange * 0.04;
    final finalRange = maxP - minP;

    double priceToY(double price) {
      return mainChartHeight - ((price - minP) / finalRange * mainChartHeight);
    }

    final candleStep = chartWidth / visibleCandles.length;
    final bodyWidth = (candleStep * 0.72).clamp(1.5, 16.0);

    // 2. Draw Gridlines & Price Scales
    final gridPaint = Paint()
      ..color = gridColor
      ..strokeWidth = 1.0;

    final borderPaint = Paint()
      ..color = borderColor
      ..strokeWidth = 1.0;

    final textStyle = TextStyle(
      color: textColor,
      fontSize: 10,
      fontFamily: 'monospace',
    );

    // Horizontal Price Grid (5 levels)
    const priceSteps = 5;
    for (int i = 0; i <= priceSteps; i++) {
      final y = mainChartHeight * (i / priceSteps);
      canvas.drawLine(Offset(0, y), Offset(chartWidth, y), gridPaint);

      final priceVal = maxP - (finalRange * (i / priceSteps));
      final tp = TextPainter(
        text: TextSpan(text: priceVal.toStringAsFixed(1), style: textStyle),
        textDirection: TextDirection.ltr,
      )..layout();
      tp.paint(canvas, Offset(chartWidth + 6, y - tp.height / 2));
    }

    // Vertical Divider separating canvas and price scale
    canvas.drawLine(
      Offset(chartWidth, 0),
      Offset(chartWidth, size.height - dateScaleHeight),
      borderPaint,
    );

    // Main Chart Bottom border
    canvas.drawLine(
      Offset(0, mainChartHeight),
      Offset(chartWidth, mainChartHeight),
      borderPaint,
    );

    // 3. Draw Volume Histogram (Bottom 22% of Main Chart)
    if (showVolume && maxVol > 0) {
      final volMaxH = mainChartHeight * 0.22;
      for (int i = 0; i < visibleCandles.length; i++) {
        final c = visibleCandles[i];
        final xCenter = (i + 0.5) * candleStep;
        final vRatio = (c.volume / maxVol).clamp(0.0, 1.0);
        final vHeight = vRatio * volMaxH;
        final vTop = mainChartHeight - vHeight;

        final vPaint = Paint()
          ..color = (c.isBullish ? bullColor : bearColor).withValues(alpha: 0.35)
          ..style = PaintingStyle.fill;

        canvas.drawRect(
          Rect.fromCenter(
            center: Offset(xCenter, vTop + vHeight / 2),
            width: bodyWidth,
            height: vHeight.clamp(1.0, volMaxH),
          ),
          vPaint,
        );
      }
    }

    // 4. Draw Candlesticks or Line Chart
    if (isLineChart) {
      final linePath = Path();
      for (int i = 0; i < visibleCandles.length; i++) {
        final c = visibleCandles[i];
        final x = (i + 0.5) * candleStep;
        final y = priceToY(c.close);
        if (i == 0) {
          linePath.moveTo(x, y);
        } else {
          linePath.lineTo(x, y);
        }
      }
      canvas.drawPath(
        linePath,
        Paint()
          ..color = const Color(0xFF3B82F6)
          ..strokeWidth = 2.0
          ..style = PaintingStyle.stroke,
      );
    } else {
      // Candlestick rendering
      for (int i = 0; i < visibleCandles.length; i++) {
        final c = visibleCandles[i];
        final xCenter = (i + 0.5) * candleStep;
        final yOpen = priceToY(c.open);
        final yClose = priceToY(c.close);
        final yHigh = priceToY(c.high);
        final yLow = priceToY(c.low);

        final isUp = c.isBullish;
        final candleColor = isUp ? bullColor : bearColor;

        // Wick
        final wickPaint = Paint()
          ..color = candleColor
          ..strokeWidth = 1.2
          ..style = PaintingStyle.stroke;

        canvas.drawLine(Offset(xCenter, yHigh), Offset(xCenter, yLow), wickPaint);

        // Body
        final bodyTop = yOpen < yClose ? yOpen : yClose;
        final bodyH = (yClose - yOpen).abs().clamp(1.5, double.infinity);

        final bodyPaint = Paint()
          ..color = candleColor
          ..style = PaintingStyle.fill;

        canvas.drawRect(
          Rect.fromLTWH(xCenter - bodyWidth / 2, bodyTop, bodyWidth, bodyH),
          bodyPaint,
        );
      }
    }

    // 5. Draw Overlaid Moving Averages (EMA 5, 13, 26, SMA 200)
    void drawMaLine(double Function(ChartIndicatorModel ind) getVal, Color color, double strokeW) {
      final path = Path();
      bool started = false;

      for (int i = 0; i < visibleCandles.length; i++) {
        final c = visibleCandles[i];
        final dtStr = DateFormat('yyyy-MM-dd').format(c.date);
        final ind = indicatorMap[dtStr];
        if (ind == null) continue;

        final val = getVal(ind);
        if (val <= 0) continue;

        final x = (i + 0.5) * candleStep;
        final y = priceToY(val);

        if (!started) {
          path.moveTo(x, y);
          started = true;
        } else {
          path.lineTo(x, y);
        }
      }

      if (started) {
        canvas.drawPath(
          path,
          Paint()
            ..color = color
            ..strokeWidth = strokeW
            ..style = PaintingStyle.stroke
            ..strokeCap = StrokeCap.round
            ..strokeJoin = StrokeJoin.round,
        );
      }
    }

    if (showEma5) drawMaLine((i) => i.ema5, ema5Color, 1.8);
    if (showEma13) drawMaLine((i) => i.ema13, ema13Color, 1.8);
    if (showEma26) drawMaLine((i) => i.ema26, ema26Color, 1.8);
    if (showSma200) drawMaLine((i) => i.sma200, sma200Color, 2.0);

    // 6. Draw RSI Sub-Chart (if enabled)
    if (showRsi) {
      // Sub-chart background & borders
      canvas.drawRect(
        Rect.fromLTWH(0, rsiTop, chartWidth, rsiHeight),
        Paint()..color = isDark ? const Color(0xFF0F172A) : const Color(0xFFF8FAFC),
      );
      canvas.drawLine(Offset(0, rsiTop), Offset(chartWidth, rsiTop), borderPaint);
      canvas.drawLine(Offset(0, rsiTop + rsiHeight), Offset(chartWidth, rsiTop + rsiHeight), borderPaint);

      double rsiToY(double rsi) {
        return rsiTop + rsiHeight - ((rsi.clamp(0.0, 100.0) / 100.0) * rsiHeight);
      }

      // 70 (Overbought) & 30 (Oversold) reference lines
      final rsi70Y = rsiToY(70);
      final rsi30Y = rsiToY(30);

      final dashed70 = Paint()
        ..color = const Color(0xFFEF5350).withValues(alpha: 0.6)
        ..strokeWidth = 1.0;
      final dashed30 = Paint()
        ..color = const Color(0xFF10B981).withValues(alpha: 0.6)
        ..strokeWidth = 1.0;

      // Draw dashed reference lines
      const dashW = 4.0;
      const spaceW = 3.0;
      for (double dx = 0; dx < chartWidth; dx += (dashW + spaceW)) {
        canvas.drawLine(Offset(dx, rsi70Y), Offset((dx + dashW).clamp(0, chartWidth), rsi70Y), dashed70);
        canvas.drawLine(Offset(dx, rsi30Y), Offset((dx + dashW).clamp(0, chartWidth), rsi30Y), dashed30);
      }

      // 70 & 30 labels on right scale
      final tp70 = TextPainter(
        text: TextSpan(text: '70', style: TextStyle(color: const Color(0xFFEF5350), fontSize: 9, fontFamily: 'monospace')),
        textDirection: TextDirection.ltr,
      )..layout();
      tp70.paint(canvas, Offset(chartWidth + 6, rsi70Y - tp70.height / 2));

      final tp30 = TextPainter(
        text: TextSpan(text: '30', style: TextStyle(color: const Color(0xFF10B981), fontSize: 9, fontFamily: 'monospace')),
        textDirection: TextDirection.ltr,
      )..layout();
      tp30.paint(canvas, Offset(chartWidth + 6, rsi30Y - tp30.height / 2));

      // Plot RSI Line
      final rsiPath = Path();
      bool rsiStarted = false;

      for (int i = 0; i < visibleCandles.length; i++) {
        final c = visibleCandles[i];
        final dtStr = DateFormat('yyyy-MM-dd').format(c.date);
        final ind = indicatorMap[dtStr];
        if (ind == null || ind.rsi < 0) continue;

        final x = (i + 0.5) * candleStep;
        final y = rsiToY(ind.rsi);

        if (!rsiStarted) {
          rsiPath.moveTo(x, y);
          rsiStarted = true;
        } else {
          rsiPath.lineTo(x, y);
        }
      }

      if (rsiStarted) {
        canvas.drawPath(
          rsiPath,
          Paint()
            ..color = rsiColor
            ..strokeWidth = 1.8
            ..style = PaintingStyle.stroke,
        );
      }

      // RSI Label
      final rsiLabel = TextPainter(
        text: TextSpan(
          text: 'RSI(14)',
          style: TextStyle(color: rsiColor, fontSize: 10, fontWeight: FontWeight.bold),
        ),
        textDirection: TextDirection.ltr,
      )..layout();
      rsiLabel.paint(canvas, Offset(6, rsiTop + 4));
    }

    // 7. Date Scale along the bottom
    final dateY = size.height - dateScaleHeight + 4;
    final dateStepInterval = (visibleCandles.length / 4).clamp(1, visibleCandles.length).toInt();

    for (int i = 0; i < visibleCandles.length; i += dateStepInterval) {
      final c = visibleCandles[i];
      final x = (i + 0.5) * candleStep;
      final dateText = DateFormat('dd MMM').format(c.date);

      final tpDate = TextPainter(
        text: TextSpan(text: dateText, style: textStyle),
        textDirection: TextDirection.ltr,
      )..layout();

      final textX = (x - tpDate.width / 2).clamp(0.0, chartWidth - tpDate.width);
      tpDate.paint(canvas, Offset(textX, dateY));
    }

    // 8. Crosshair Inspection Overlay (if active)
    if (crosshairPosition != null) {
      final touchX = crosshairPosition!.dx.clamp(0.0, chartWidth);
      final touchY = crosshairPosition!.dy.clamp(0.0, size.height - dateScaleHeight);

      // Identify closest candle
      final candleIdx = (touchX / candleStep).floor().clamp(0, visibleCandles.length - 1);
      final selectedCandle = visibleCandles[candleIdx];
      final dtStr = DateFormat('yyyy-MM-dd').format(selectedCandle.date);
      final selectedIndicator = indicatorMap[dtStr];

      final candleCenterX = (candleIdx + 0.5) * candleStep;

      final crosshairPaint = Paint()
        ..color = isDark ? Colors.white60 : Colors.black45
        ..strokeWidth = 1.0;

      // Vertical line
      for (double dy = 0; dy < (size.height - dateScaleHeight); dy += 6) {
        canvas.drawLine(Offset(candleCenterX, dy), Offset(candleCenterX, dy + 3), crosshairPaint);
      }

      // Horizontal line (if touched in main chart)
      if (touchY <= mainChartHeight) {
        for (double dx = 0; dx < chartWidth; dx += 6) {
          canvas.drawLine(Offset(dx, touchY), Offset(dx + 3, touchY), crosshairPaint);
        }

        // Draw Price Pill on right scale
        final touchPrice = maxP - (touchY / mainChartHeight * finalRange);
        final pricePillText = TextPainter(
          text: TextSpan(
            text: touchPrice.toStringAsFixed(2),
            style: const TextStyle(color: Colors.white, fontSize: 10, fontWeight: FontWeight.bold),
          ),
          textDirection: TextDirection.ltr,
        )..layout();

        final pillRect = Rect.fromLTWH(
          chartWidth + 2,
          touchY - pricePillText.height / 2 - 2,
          pricePillText.width + 8,
          pricePillText.height + 4,
        );
        canvas.drawRRect(
          RRect.fromRectAndRadius(pillRect, const Radius.circular(4)),
          Paint()..color = const Color(0xFF1E293B),
        );
        pricePillText.paint(canvas, Offset(chartWidth + 6, touchY - pricePillText.height / 2));
      }

      // Notify parent widget of inspected candle
      WidgetsBinding.instance.addPostFrameCallback((_) {
        onCrosshairUpdated?.call(selectedCandle, selectedIndicator);
      });
    }
  }

  @override
  bool shouldRepaint(covariant TechnicalChartPainter oldDelegate) {
    return oldDelegate.candles != candles ||
        oldDelegate.startIndex != startIndex ||
        oldDelegate.visibleCount != visibleCount ||
        oldDelegate.isDark != isDark ||
        oldDelegate.isLineChart != isLineChart ||
        oldDelegate.showEma5 != showEma5 ||
        oldDelegate.showEma13 != showEma13 ||
        oldDelegate.showEma26 != showEma26 ||
        oldDelegate.showSma200 != showSma200 ||
        oldDelegate.showVolume != showVolume ||
        oldDelegate.showRsi != showRsi ||
        oldDelegate.crosshairPosition != crosshairPosition;
  }
}
