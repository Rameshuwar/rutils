class BMIResult {
  final double bmi;
  final String category;

  BMIResult({required this.bmi, required this.category});

  factory BMIResult.fromJson(Map<String, dynamic> json) {
    return BMIResult(
      bmi: (json['bmi'] is num) ? (json['bmi'] as num).toDouble() : 0.0,
      category: json['category'] ?? '',
    );
  }
}

class AgeResultModel {
  final int years;
  final int months;
  final int days;
  final NextBirthdayModel nextBirthday;
  final AgeSummaryModel summary;

  AgeResultModel({
    required this.years,
    required this.months,
    required this.days,
    required this.nextBirthday,
    required this.summary,
  });

  factory AgeResultModel.fromJson(Map<String, dynamic> json) {
    final age = json['age'] ?? {};
    return AgeResultModel(
      years: age['years'] ?? 0,
      months: age['months'] ?? 0,
      days: age['days'] ?? 0,
      nextBirthday: NextBirthdayModel.fromJson(json['nextBirthday'] ?? {}),
      summary: AgeSummaryModel.fromJson(json['summary'] ?? {}),
    );
  }
}

class NextBirthdayModel {
  final String date;
  final int day;
  final String dayOfWeek;
  final int daysRemaining;
  final int month;
  final int monthsRemaining;

  NextBirthdayModel({
    required this.date,
    required this.day,
    required this.dayOfWeek,
    required this.daysRemaining,
    required this.month,
    required this.monthsRemaining,
  });

  factory NextBirthdayModel.fromJson(Map<String, dynamic> json) {
    return NextBirthdayModel(
      date: json['date'] ?? '',
      day: json['day'] ?? 0,
      dayOfWeek: json['dayOfWeek'] ?? '',
      daysRemaining: json['daysRemaining'] ?? 0,
      month: json['month'] ?? 0,
      monthsRemaining: json['monthsRemaining'] ?? 0,
    );
  }
}

class AgeSummaryModel {
  final int years;
  final int months;
  final int weeks;
  final int days;
  final int hours;
  final int minutes;

  AgeSummaryModel({
    required this.years,
    required this.months,
    required this.weeks,
    required this.days,
    required this.hours,
    required this.minutes,
  });

  factory AgeSummaryModel.fromJson(Map<String, dynamic> json) {
    return AgeSummaryModel(
      years: json['years'] ?? 0,
      months: json['months'] ?? 0,
      weeks: json['weeks'] ?? 0,
      days: json['days'] ?? 0,
      hours: json['hours'] ?? 0,
      minutes: json['minutes'] ?? 0,
    );
  }
}

class PercentageResultModel {
  final String operation;
  final double result;
  final String formatted;
  final Map<String, dynamic>? extra;
  final List<String> steps;

  PercentageResultModel({
    required this.operation,
    required this.result,
    required this.formatted,
    this.extra,
    required this.steps,
  });

  factory PercentageResultModel.fromJson(Map<String, dynamic> json) {
    return PercentageResultModel(
      operation: json['operation'] ?? '',
      result: (json['result'] is num) ? (json['result'] as num).toDouble() : 0.0,
      formatted: json['formatted']?.toString() ?? json['result']?.toString() ?? '',
      extra: json['extra'] is Map<String, dynamic> ? json['extra'] : null,
      steps: (json['steps'] as List<dynamic>?)?.map((s) => s.toString()).toList() ?? [],
    );
  }
}

class EMIResultModel {
  final double emi;
  final double principal;
  final int tenureMonths;
  final double monthlyRatePercent;
  final double totalInterest;
  final double totalPayment;
  final double principalPercent;
  final double interestPercent;
  final List<AmortizationRowModel> amortization;

  EMIResultModel({
    required this.emi,
    required this.principal,
    required this.tenureMonths,
    required this.monthlyRatePercent,
    required this.totalInterest,
    required this.totalPayment,
    required this.principalPercent,
    required this.interestPercent,
    required this.amortization,
  });

  factory EMIResultModel.fromJson(Map<String, dynamic> json) {
    final breakdown = json['breakdown'] ?? {};
    final amortList = (json['amortization'] as List<dynamic>?)
            ?.map((row) => AmortizationRowModel.fromJson(row))
            .toList() ??
        [];

    return EMIResultModel(
      emi: (json['emi'] is num) ? (json['emi'] as num).toDouble() : 0.0,
      principal: (json['principal'] is num) ? (json['principal'] as num).toDouble() : 0.0,
      tenureMonths: json['tenureMonths'] ?? 0,
      monthlyRatePercent: (json['monthlyRatePercent'] is num)
          ? (json['monthlyRatePercent'] as num).toDouble()
          : 0.0,
      totalInterest:
          (json['totalInterest'] is num) ? (json['totalInterest'] as num).toDouble() : 0.0,
      totalPayment:
          (json['totalPayment'] is num) ? (json['totalPayment'] as num).toDouble() : 0.0,
      principalPercent: (breakdown['principalPercent'] is num)
          ? (breakdown['principalPercent'] as num).toDouble()
          : 0.0,
      interestPercent: (breakdown['interestPercent'] is num)
          ? (breakdown['interestPercent'] as num).toDouble()
          : 0.0,
      amortization: amortList,
    );
  }
}

class AmortizationRowModel {
  final int month;
  final double openingBalance;
  final double principalPaid;
  final double interestPaid;
  final double totalPaid;
  final double closingBalance;

  AmortizationRowModel({
    required this.month,
    required this.openingBalance,
    required this.principalPaid,
    required this.interestPaid,
    required this.totalPaid,
    required this.closingBalance,
  });

  factory AmortizationRowModel.fromJson(Map<String, dynamic> json) {
    return AmortizationRowModel(
      month: json['month'] ?? 0,
      openingBalance: (json['openingBalance'] is num) ? (json['openingBalance'] as num).toDouble() : 0.0,
      principalPaid: (json['principalPaid'] is num) ? (json['principalPaid'] as num).toDouble() : 0.0,
      interestPaid: (json['interestPaid'] is num) ? (json['interestPaid'] as num).toDouble() : 0.0,
      totalPaid: (json['totalPaid'] is num) ? (json['totalPaid'] as num).toDouble() : 0.0,
      closingBalance: (json['closingBalance'] is num) ? (json['closingBalance'] as num).toDouble() : 0.0,
    );
  }
}

class TaxResultModel {
  final double grossAmount;
  final double netAmount;
  final double taxAmount;
  final double taxRate;
  final String formatted;
  final List<String> steps;
  final Map<String, dynamic>? extra;

  TaxResultModel({
    required this.grossAmount,
    required this.netAmount,
    required this.taxAmount,
    required this.taxRate,
    required this.formatted,
    required this.steps,
    this.extra,
  });

  factory TaxResultModel.fromJson(Map<String, dynamic> json) {
    return TaxResultModel(
      grossAmount: (json['grossAmount'] is num) ? (json['grossAmount'] as num).toDouble() : 0.0,
      netAmount: (json['netAmount'] is num) ? (json['netAmount'] as num).toDouble() : 0.0,
      taxAmount: (json['taxAmount'] is num) ? (json['taxAmount'] as num).toDouble() : 0.0,
      taxRate: (json['taxRate'] is num) ? (json['taxRate'] as num).toDouble() : 0.0,
      formatted: json['formatted']?.toString() ?? '',
      steps: (json['steps'] as List<dynamic>?)?.map((s) => s.toString()).toList() ?? [],
      extra: json['extra'] is Map<String, dynamic> ? json['extra'] : null,
    );
  }
}
