class NiftyCompanyModel {
  final String companyName;
  final String industry;
  final String symbol;
  final String series;
  final String isin;

  NiftyCompanyModel({
    required this.companyName,
    required this.industry,
    required this.symbol,
    required this.series,
    required this.isin,
  });

  factory NiftyCompanyModel.fromJson(Map<String, dynamic> json) {
    return NiftyCompanyModel(
      companyName: json['company_name'] ?? '',
      industry: json['industry'] ?? 'General',
      symbol: json['symbol'] ?? '',
      series: json['series'] ?? 'EQ',
      isin: json['isin'] ?? '',
    );
  }
}

class NiftyResponseModel {
  final bool success;
  final String updatedAt;
  final int count;
  final List<NiftyCompanyModel> companies;
  final String? error;

  NiftyResponseModel({
    required this.success,
    required this.updatedAt,
    required this.count,
    required this.companies,
    this.error,
  });

  factory NiftyResponseModel.fromJson(Map<String, dynamic> json) {
    final list = (json['companies'] as List<dynamic>?)
            ?.map((c) => NiftyCompanyModel.fromJson(c as Map<String, dynamic>))
            .toList() ??
        [];

    return NiftyResponseModel(
      success: json['success'] ?? false,
      updatedAt: json['updated_at'] ?? '',
      count: json['count'] ?? list.length,
      companies: list,
      error: json['error'],
    );
  }
}
