import '../../../core/constants/api_constants.dart';
import '../../../core/network/api_client.dart';
import '../models/calculation_models.dart';

class CalculationService {
  final ApiClient _api = ApiClient();

  Future<BMIResult> calculateBMI({
    required double height,
    required String heightUnit,
    required double weight,
    required String weightUnit,
  }) async {
    final data = await _api.post(
      ApiConstants.calculateBmi,
      data: {
        'height': height,
        'heightUnit': heightUnit,
        'weight': weight,
        'weightUnit': weightUnit,
      },
    );
    return BMIResult.fromJson(data);
  }

  Future<AgeResultModel> calculateAge({
    required String dob,
    required String today,
  }) async {
    final data = await _api.post(
      ApiConstants.calculateAge,
      data: {
        'dob': dob,
        'today': today,
      },
    );
    return AgeResultModel.fromJson(data);
  }

  Future<PercentageResultModel> calculatePercentage({
    required String operation,
    required double value1,
    double? value2,
    double? value3,
  }) async {
    final payload = <String, dynamic>{
      'operation': operation,
      'value1': value1,
    };
    if (value2 != null) payload['value2'] = value2;
    if (value3 != null) payload['value3'] = value3;

    final data = await _api.post(
      ApiConstants.calculatePercentage,
      data: payload,
    );
    return PercentageResultModel.fromJson(data);
  }

  Future<EMIResultModel> calculateEMI({
    required double principal,
    required double annualInterestRate,
    required double tenure,
    required String tenureUnit,
  }) async {
    final data = await _api.post(
      ApiConstants.calculateEmi,
      data: {
        'principal': principal,
        'annualInterestRate': annualInterestRate,
        'tenure': tenure,
        'tenureUnit': tenureUnit,
      },
    );
    return EMIResultModel.fromJson(data);
  }

  Future<TaxResultModel> calculateTax({
    required String mode,
    double? amount,
    double? rate,
    String? taxType,
    double? net,
    double? gross,
    double? taxPaid,
    double? income,
    String? regime,
    dynamic slabs,
  }) async {
    final payload = <String, dynamic>{'mode': mode};
    if (amount != null) payload['amount'] = amount;
    if (rate != null) payload['taxRate'] = rate;
    if (taxType != null) payload['taxType'] = taxType;
    if (net != null) payload['netAmount'] = net;
    if (gross != null) payload['grossAmount'] = gross;
    if (taxPaid != null) payload['taxPaid'] = taxPaid;
    if (income != null) payload['income'] = income;
    if (regime != null) payload['regime'] = regime;
    if (slabs != null) payload['slabs'] = slabs;

    final data = await _api.post(
      ApiConstants.calculateTax,
      data: payload,
    );
    return TaxResultModel.fromJson(data);
  }
}
