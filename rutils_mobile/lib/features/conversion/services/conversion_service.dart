import 'dart:typed_data';
import 'package:dio/dio.dart';
import '../../../core/constants/api_constants.dart';
import '../../../core/network/api_client.dart';
import '../models/conversion_models.dart';

class ConversionService {
  final ApiClient _api = ApiClient();

  Future<List<int>> convertFile({
    required Uint8List fileBytes,
    required String fileName,
    required String fromType,
    required String toType,
    void Function(int sent, int total)? onProgress,
  }) async {
    final formData = FormData.fromMap({
      'file': MultipartFile.fromBytes(fileBytes, filename: fileName),
      'fromType': fromType.toLowerCase(),
      'toType': toType.toLowerCase(),
    });

    final response = await _api.postMultipartBytes(
      ApiConstants.convertFile,
      formData: formData,
      onSendProgress: onProgress,
    );
    return response.data ?? [];
  }

  Future<List<int>> convertPdfSize({
    required Uint8List fileBytes,
    required String fileName,
    required String conversionType,
    required String dataType,
    required double targetSize,
    void Function(int sent, int total)? onProgress,
  }) async {
    final formData = FormData.fromMap({
      'file': MultipartFile.fromBytes(fileBytes, filename: fileName),
      'conversionType': conversionType,
      'dataType': dataType,
      'targetSize': targetSize,
    });

    final response = await _api.postMultipartBytes(
      ApiConstants.convertPdfSize,
      formData: formData,
      onSendProgress: onProgress,
    );
    return response.data ?? [];
  }

  Future<MeasurementResult> convertMeasurement({
    required String category,
    required String fromUnit,
    required String toUnit,
    required double value,
  }) async {
    final data = await _api.post(
      ApiConstants.convertMeasurement,
      data: {
        'category': category,
        'fromUnit': fromUnit,
        'toUnit': toUnit,
        'value': value,
      },
    );
    return MeasurementResult.fromJson(data);
  }

  Future<TimeConversionResult> convertTime({
    required int year,
    required int month,
    required int day,
    required int hour,
    required int minute,
    required int second,
    required String sourceTz,
    required String destTz,
    String ambiguousPolicy = 'earlier',
    String nonExistentPolicy = 'shift_forward',
  }) async {
    final data = await _api.post(
      ApiConstants.convertTime,
      data: {
        'year': year,
        'month': month,
        'day': day,
        'hour': hour,
        'minute': minute,
        'second': second,
        'source_tz': sourceTz,
        'dest_tz': destTz,
        'ambiguous_policy': ambiguousPolicy,
        'non_existent_policy': nonExistentPolicy,
      },
    );
    return TimeConversionResult.fromJson(data);
  }

  Future<RailwayResult> convertRailway({
    required String direction,
    required int hour,
    required int minute,
    String? ampm,
  }) async {
    final payload = <String, dynamic>{
      'direction': direction,
      'hour': hour,
      'minute': minute,
    };
    if (ampm != null) payload['ampm'] = ampm;

    final data = await _api.post(
      ApiConstants.convertRailway,
      data: payload,
    );
    return RailwayResult.fromJson(data);
  }

  Future<NumeralResult> convertNumeral({
    required String value,
    required String fromBase,
    required String toBase,
  }) async {
    final data = await _api.post(
      ApiConstants.convertNumeral,
      data: {
        'value': value,
        'fromBase': fromBase,
        'toBase': toBase,
      },
    );
    return NumeralResult.fromJson(data);
  }
}
