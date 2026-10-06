class MeasurementResult {
  final double result;
  MeasurementResult({required this.result});

  factory MeasurementResult.fromJson(Map<String, dynamic> json) {
    final res = json['result'];
    return MeasurementResult(
      result: (res is num) ? res.toDouble() : double.tryParse(res.toString()) ?? 0.0,
    );
  }
}

class TimeConversionResult {
  final String sourceTimeLocal;
  final String sourceTimeUtc;
  final String sourceZoneName;
  final int sourceOffset;
  final String destTimeLocal;
  final String destTimeUtc;
  final String destZoneName;
  final int destOffset;
  final bool isNextDay;
  final bool isPrevDay;
  final String? warning;

  TimeConversionResult({
    required this.sourceTimeLocal,
    required this.sourceTimeUtc,
    required this.sourceZoneName,
    required this.sourceOffset,
    required this.destTimeLocal,
    required this.destTimeUtc,
    required this.destZoneName,
    required this.destOffset,
    required this.isNextDay,
    required this.isPrevDay,
    this.warning,
  });

  factory TimeConversionResult.fromJson(Map<String, dynamic> json) {
    return TimeConversionResult(
      sourceTimeLocal: json['source_time_local'] ?? '',
      sourceTimeUtc: json['source_time_utc'] ?? '',
      sourceZoneName: json['source_zone_name'] ?? '',
      sourceOffset: json['source_offset'] ?? 0,
      destTimeLocal: json['dest_time_local'] ?? '',
      destTimeUtc: json['dest_time_utc'] ?? '',
      destZoneName: json['dest_zone_name'] ?? '',
      destOffset: json['dest_offset'] ?? 0,
      isNextDay: json['is_next_day'] ?? false,
      isPrevDay: json['is_prev_day'] ?? false,
      warning: json['warning'],
    );
  }
}

class RailwayResult {
  final int? hour12;
  final int? hour24;
  final int minute;
  final String? ampm;

  RailwayResult({
    this.hour12,
    this.hour24,
    required this.minute,
    this.ampm,
  });

  factory RailwayResult.fromJson(Map<String, dynamic> json) {
    return RailwayResult(
      hour12: json['hour_12'],
      hour24: json['hour_24'],
      minute: json['minute'] ?? 0,
      ampm: json['ampm'],
    );
  }
}

class NumeralResult {
  final String result;
  NumeralResult({required this.result});

  factory NumeralResult.fromJson(Map<String, dynamic> json) {
    return NumeralResult(result: json['result']?.toString() ?? '');
  }
}
