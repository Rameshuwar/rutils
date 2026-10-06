import '../../../core/constants/api_constants.dart';
import '../../../core/network/api_client.dart';
import '../models/chart_models.dart';

class ChartService {
  final ApiClient _client = ApiClient();

  Future<List<ChartCompanyModel>> fetchCompanies() async {
    final res = await _client.get(ApiConstants.chartCompanies);
    if (res is Map && res['companies'] is List) {
      final list = res['companies'] as List;
      return list.map((c) => ChartCompanyModel.fromJson(c as Map<String, dynamic>)).toList();
    }
    return [];
  }

  Future<ChartDataResponse> fetchChartData(String symbol) async {
    final res = await _client.get(
      ApiConstants.chartData,
      queryParameters: {'symbol': symbol},
    );
    if (res is Map<String, dynamic>) {
      return ChartDataResponse.fromJson(res);
    }
    throw Exception('Invalid chart data response from server');
  }
}
