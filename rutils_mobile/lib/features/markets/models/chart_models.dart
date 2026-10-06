class ChartCompanyModel {
  final String symbol;
  final String displayName;
  final String? industry;
  final String? symbolToken;
  final String exchange;
  final String interval;
  final int candleCount;
  final int adviceCount;
  final String? latestCandleDate;

  ChartCompanyModel({
    required this.symbol,
    required this.displayName,
    this.industry,
    this.symbolToken,
    required this.exchange,
    required this.interval,
    required this.candleCount,
    required this.adviceCount,
    this.latestCandleDate,
  });

  factory ChartCompanyModel.fromJson(Map<String, dynamic> json) {
    return ChartCompanyModel(
      symbol: (json['symbol'] ?? '').toString(),
      displayName: (json['displayName'] ?? json['symbol'] ?? '').toString(),
      industry: json['industry']?.toString(),
      symbolToken: json['symbolToken']?.toString(),
      exchange: (json['exchange'] ?? 'NSE').toString(),
      interval: (json['interval'] ?? 'ONE_DAY').toString(),
      candleCount: (json['candleCount'] as num?)?.toInt() ?? 0,
      adviceCount: (json['adviceCount'] as num?)?.toInt() ?? 0,
      latestCandleDate: json['latestCandleDate']?.toString(),
    );
  }
}

class ChartCandleModel {
  final DateTime date;
  final double open;
  final double high;
  final double low;
  final double close;
  final double volume;

  ChartCandleModel({
    required this.date,
    required this.open,
    required this.high,
    required this.low,
    required this.close,
    required this.volume,
  });

  factory ChartCandleModel.fromJson(Map<String, dynamic> json) {
    final rawDate = (json['date'] ?? '').toString();
    DateTime dt;
    try {
      dt = DateTime.parse(rawDate);
    } catch (_) {
      dt = DateTime.now();
    }

    return ChartCandleModel(
      date: dt,
      open: (json['open'] as num?)?.toDouble() ?? 0.0,
      high: (json['high'] as num?)?.toDouble() ?? 0.0,
      low: (json['low'] as num?)?.toDouble() ?? 0.0,
      close: (json['close'] as num?)?.toDouble() ?? 0.0,
      volume: (json['volume'] as num?)?.toDouble() ?? 0.0,
    );
  }

  bool get isBullish => close >= open;
}

class ChartIndicatorModel {
  final DateTime date;
  final double close;
  final double ema5;
  final double ema13;
  final double ema26;
  final double sma200;
  final double rsi;

  ChartIndicatorModel({
    required this.date,
    required this.close,
    required this.ema5,
    required this.ema13,
    required this.ema26,
    required this.sma200,
    required this.rsi,
  });

  factory ChartIndicatorModel.fromJson(Map<String, dynamic> json) {
    final rawDate = (json['date'] ?? '').toString();
    DateTime dt;
    try {
      dt = DateTime.parse(rawDate);
    } catch (_) {
      dt = DateTime.now();
    }

    return ChartIndicatorModel(
      date: dt,
      close: (json['close'] as num?)?.toDouble() ?? 0.0,
      ema5: (json['ema5'] as num?)?.toDouble() ?? 0.0,
      ema13: (json['ema13'] as num?)?.toDouble() ?? 0.0,
      ema26: (json['ema26'] as num?)?.toDouble() ?? 0.0,
      sma200: (json['sma200'] as num?)?.toDouble() ?? 0.0,
      rsi: (json['rsi'] as num?)?.toDouble() ?? 0.0,
    );
  }
}

class ChartAdviceModel {
  final String advice;
  final double targetPrice;
  final double sma200Support;
  final String date;

  ChartAdviceModel({
    required this.advice,
    required this.targetPrice,
    required this.sma200Support,
    required this.date,
  });

  factory ChartAdviceModel.fromJson(Map<String, dynamic> json) {
    return ChartAdviceModel(
      advice: (json['advice'] ?? 'HOLD').toString().toUpperCase(),
      targetPrice: (json['targetPrice'] as num?)?.toDouble() ?? 0.0,
      sma200Support: (json['sma200Support'] as num?)?.toDouble() ?? 0.0,
      date: (json['date'] ?? '').toString(),
    );
  }
}

class ChartDataResponse {
  final bool success;
  final String symbol;
  final String companyName;
  final List<ChartCandleModel> candles;
  final List<ChartIndicatorModel> indicators;
  final List<ChartAdviceModel> advice;

  ChartDataResponse({
    required this.success,
    required this.symbol,
    required this.companyName,
    required this.candles,
    required this.indicators,
    required this.advice,
  });

  factory ChartDataResponse.fromJson(Map<String, dynamic> json) {
    final meta = json['metadata'] as Map<String, dynamic>? ?? {};
    final rawCandles = json['candles'] as List? ?? [];
    final rawIndicators = json['indicators'] as List? ?? [];
    final rawAdvice = json['advice'] as List? ?? [];

    return ChartDataResponse(
      success: json['success'] == true,
      symbol: (json['symbol'] ?? '').toString(),
      companyName: (meta['companyName'] ?? json['symbol'] ?? '').toString(),
      candles: rawCandles.map((c) => ChartCandleModel.fromJson(c as Map<String, dynamic>)).toList(),
      indicators: rawIndicators.map((i) => ChartIndicatorModel.fromJson(i as Map<String, dynamic>)).toList(),
      advice: rawAdvice.map((a) => ChartAdviceModel.fromJson(a as Map<String, dynamic>)).toList(),
    );
  }
}
