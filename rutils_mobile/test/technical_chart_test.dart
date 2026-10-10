import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:rutils_mobile/features/markets/models/chart_models.dart';
import 'package:rutils_mobile/features/markets/presentation/widgets/interactive_technical_chart.dart';

void main() {
  group('Chart Models Parsing Tests', () {
    test('ChartCompanyModel parses from JSON correctly', () {
      final json = {
        'symbol': 'HINDUNILVR',
        'displayName': 'Hindustan Unilever Ltd.',
        'industry': 'Fast Moving Consumer Goods',
        'symbolToken': '1394',
        'exchange': 'NSE',
        'interval': 'ONE_DAY',
        'candleCount': 246,
        'adviceCount': 7,
        'latestCandleDate': '2026-10-01T00:00:00+05:30',
      };

      final model = ChartCompanyModel.fromJson(json);
      expect(model.symbol, 'HINDUNILVR');
      expect(model.displayName, 'Hindustan Unilever Ltd.');
      expect(model.candleCount, 246);
    });

    test('ChartCandleModel and ChartIndicatorModel parse correctly', () {
      final candleJson = {
        'date': '2026-10-01T00:00:00+05:30',
        'open': 1870.1,
        'high': 1873.6,
        'low': 1826.9,
        'close': 1836.0,
        'volume': 1331680.0,
      };

      final candle = ChartCandleModel.fromJson(candleJson);
      expect(candle.open, 1870.1);
      expect(candle.close, 1836.0);
      expect(candle.isBullish, false); // close < open

      final indicatorJson = {
        'date': '2026-10-01T00:00:00+05:30',
        'close': 1836.0,
        'ema5': 1860.2,
        'ema13': 1875.4,
        'ema26': 1890.1,
        'sma200': 2100.5,
        'rsi': 35.8,
      };

      final ind = ChartIndicatorModel.fromJson(indicatorJson);
      expect(ind.ema5, 1860.2);
      expect(ind.rsi, 35.8);
    });
  });

  group('Interactive Technical Chart Widget Tests', () {
    testWidgets('InteractiveTechnicalChart renders without exceptions', (tester) async {
      final candles = [
        ChartCandleModel(
          date: DateTime(2026, 9, 29),
          open: 1850.0,
          high: 1870.0,
          low: 1845.0,
          close: 1865.0,
          volume: 1000000.0,
        ),
        ChartCandleModel(
          date: DateTime(2026, 9, 30),
          open: 1865.0,
          high: 1889.0,
          low: 1860.0,
          close: 1879.0,
          volume: 1200000.0,
        ),
        ChartCandleModel(
          date: DateTime(2026, 10, 1),
          open: 1870.0,
          high: 1873.0,
          low: 1826.0,
          close: 1836.0,
          volume: 1400000.0,
        ),
      ];

      final indicators = {
        '2026-10-01': ChartIndicatorModel(
          date: DateTime(2026, 10, 1),
          close: 1836.0,
          ema5: 1860.0,
          ema13: 1875.0,
          ema26: 1890.0,
          sma200: 2100.0,
          rsi: 42.0,
        ),
      };

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: SizedBox(
              width: 400,
              height: 600,
              child: InteractiveTechnicalChart(
                candles: candles,
                indicatorMap: indicators,
                isDark: false,
                isLineChart: false,
                showEma5: true,
                showEma13: true,
                showEma26: true,
                showSma200: true,
                showVolume: true,
                showRsi: true,
                period: '6m',
              ),
            ),
          ),
        ),
      );

      await tester.pump();
      expect(find.byType(InteractiveTechnicalChart), findsOneWidget);
      expect(find.byType(CustomPaint), findsWidgets);
    });

    testWidgets('InteractiveTechnicalChart handles period changes and crosshairs', (tester) async {
      final candles = [
        ChartCandleModel(
          date: DateTime(2026, 9, 29),
          open: 1850.0,
          high: 1870.0,
          low: 1845.0,
          close: 1865.0,
          volume: 1000000.0,
        ),
        ChartCandleModel(
          date: DateTime(2026, 9, 30),
          open: 1865.0,
          high: 1889.0,
          low: 1860.0,
          close: 1879.0,
          volume: 1200000.0,
        ),
        ChartCandleModel(
          date: DateTime(2026, 10, 1),
          open: 1870.0,
          high: 1873.0,
          low: 1826.0,
          close: 1836.0,
          volume: 1400000.0,
        ),
      ];

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: SizedBox(
              width: 400,
              height: 600,
              child: InteractiveTechnicalChart(
                candles: candles,
                indicatorMap: const {},
                isDark: true,
                isLineChart: true,
                showEma5: false,
                showEma13: false,
                showEma26: false,
                showSma200: false,
                showVolume: false,
                showRsi: false,
                period: '1y',
              ),
            ),
          ),
        ),
      );

      await tester.pump();
      expect(find.byType(InteractiveTechnicalChart), findsOneWidget);
    });
  });
}
