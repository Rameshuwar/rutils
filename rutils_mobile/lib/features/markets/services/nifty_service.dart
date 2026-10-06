import '../../../core/constants/api_constants.dart';
import '../../../core/network/api_client.dart';
import '../models/nifty_company_model.dart';

class NiftyService {
  final ApiClient _api = ApiClient();

  Future<NiftyResponseModel> fetchCompanies() async {
    final data = await _api.get(ApiConstants.niftyCompanies);
    return NiftyResponseModel.fromJson(data);
  }
}
