import 'dart:convert';
import 'package:dio/dio.dart';
import 'package:shared_preferences/shared_preferences.dart';
import '../constants/api_constants.dart';
import 'api_exception.dart';

class ApiClient {
  static final ApiClient _instance = ApiClient._internal();
  factory ApiClient() => _instance;

  late Dio dio;
  String _baseUrl = ApiConstants.getDefaultBaseUrl();
  String? _authToken;

  ApiClient._internal() {
    dio = Dio(
      BaseOptions(
        baseUrl: _baseUrl,
        connectTimeout: const Duration(seconds: 30),
        receiveTimeout: const Duration(seconds: 45),
        headers: {
          'Accept': 'application/json',
        },
      ),
    );

    dio.interceptors.add(
      InterceptorsWrapper(
        onRequest: (options, handler) {
          if (_authToken != null && _authToken!.isNotEmpty) {
            options.headers['Authorization'] = 'Bearer $_authToken';
          }
          return handler.next(options);
        },
        onError: (DioException e, handler) {
          return handler.next(e);
        },
      ),
    );
  }

  String get baseUrl => _baseUrl;
  String? get authToken => _authToken;

  void setBaseUrl(String url) {
    _baseUrl = url.trim().replaceAll(RegExp(r'/+$'), '');
    dio.options.baseUrl = _baseUrl;
  }

  void setAuthToken(String? token) {
    _authToken = token;
  }

  Future<void> initFromStorage() async {
    final prefs = await SharedPreferences.getInstance();
    final savedToken = prefs.getString('rutils_token');
    if (savedToken != null && savedToken.isNotEmpty) {
      _authToken = savedToken;
    }
  }

  ApiException _handleError(dynamic error) {
    if (error is DioException) {
      final statusCode = error.response?.statusCode;
      final data = error.response?.data;

      if (data is Map) {
        final message = data['error'] ?? data['message'] ?? 'Request failed ($statusCode)';
        return ApiException(message.toString(), statusCode: statusCode, details: data);
      } else if (data is String && data.isNotEmpty) {
        try {
          final parsed = jsonDecode(data);
          if (parsed is Map && (parsed['error'] != null || parsed['message'] != null)) {
            return ApiException((parsed['error'] ?? parsed['message']).toString(), statusCode: statusCode);
          }
        } catch (_) {}
        return ApiException(data, statusCode: statusCode);
      }

      switch (error.type) {
        case DioExceptionType.connectionTimeout:
        case DioExceptionType.sendTimeout:
        case DioExceptionType.receiveTimeout:
          return ApiException('Connection timeout. Please check backend server status.', statusCode: statusCode);
        case DioExceptionType.connectionError:
          return ApiException('Cannot connect to backend server at $_baseUrl. Please verify if the server is running.', statusCode: statusCode);
        case DioExceptionType.badResponse:
          return ApiException('Server returned error code $statusCode', statusCode: statusCode);
        default:
          return ApiException(error.message ?? 'Network error occurred', statusCode: statusCode);
      }
    } else if (error is ApiException) {
      return error;
    }
    return ApiException(error.toString());
  }

  Future<dynamic> get(String path, {Map<String, dynamic>? queryParameters}) async {
    try {
      final response = await dio.get(path, queryParameters: queryParameters);
      return response.data;
    } catch (e) {
      throw _handleError(e);
    }
  }

  Future<dynamic> post(String path, {dynamic data, Map<String, dynamic>? queryParameters}) async {
    try {
      final response = await dio.post(
        path,
        data: data,
        queryParameters: queryParameters,
        options: Options(
          contentType: Headers.jsonContentType,
        ),
      );
      return response.data;
    } catch (e) {
      throw _handleError(e);
    }
  }

  Future<Response<List<int>>> postMultipartBytes(
    String path, {
    required FormData formData,
    void Function(int sent, int total)? onSendProgress,
  }) async {
    try {
      final response = await dio.post<List<int>>(
        path,
        data: formData,
        onSendProgress: onSendProgress,
        options: Options(
          responseType: ResponseType.bytes,
        ),
      );
      return response;
    } catch (e) {
      throw _handleError(e);
    }
  }

  Future<dynamic> postMultipartJson(
    String path, {
    required FormData formData,
    void Function(int sent, int total)? onSendProgress,
  }) async {
    try {
      final response = await dio.post(
        path,
        data: formData,
        onSendProgress: onSendProgress,
      );
      return response.data;
    } catch (e) {
      throw _handleError(e);
    }
  }
}
