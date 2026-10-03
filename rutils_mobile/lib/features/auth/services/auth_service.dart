import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:shared_preferences/shared_preferences.dart';
import '../../../core/constants/api_constants.dart';
import '../../../core/network/api_client.dart';
import '../models/user_model.dart';

class AuthService extends ChangeNotifier {
  static final AuthService _instance = AuthService._internal();
  factory AuthService() => _instance;

  final ApiClient _apiClient = ApiClient();
  UserModel? _currentUser;
  bool _isInitialized = false;

  AuthService._internal();

  UserModel? get currentUser => _currentUser;
  bool get isLoggedIn => _apiClient.authToken != null && _currentUser != null;
  bool get isInitialized => _isInitialized;

  static const String _keyToken = 'rutils_token';
  static const String _keyUser = 'rutils_user';

  Future<void> init() async {
    final prefs = await SharedPreferences.getInstance();
    final token = prefs.getString(_keyToken);
    final userJson = prefs.getString(_keyUser);

    if (token != null && token.isNotEmpty && userJson != null) {
      try {
        _apiClient.setAuthToken(token);
        _currentUser = UserModel.fromJson(jsonDecode(userJson));
      } catch (e) {
        await logout();
      }
    }
    _isInitialized = true;
    notifyListeners();
  }

  Future<LoginResponse> login(String email, String password) async {
    final data = await _apiClient.post(
      ApiConstants.login,
      data: {
        'email': email.trim(),
        'password': password,
      },
    );

    final loginResponse = LoginResponse.fromJson(data);
    await _saveAuth(loginResponse.token, loginResponse.user);
    return loginResponse;
  }

  Future<UserModel> register({
    required String name,
    required String email,
    required String password,
    required String confirmPassword,
  }) async {
    final data = await _apiClient.post(
      ApiConstants.register,
      data: {
        'name': name.trim(),
        'email': email.trim(),
        'password': password,
        'confirm_password': confirmPassword,
      },
    );

    return UserModel.fromJson(data);
  }

  Future<String> forgotPassword(String email) async {
    final data = await _apiClient.post(
      ApiConstants.forgotPassword,
      data: {'email': email.trim()},
    );
    return data['message'] ?? 'Temporary password sent.';
  }

  Future<String> resetPassword({
    required String newPassword,
    required String confirmPassword,
  }) async {
    final data = await _apiClient.post(
      ApiConstants.resetPassword,
      data: {
        'new_password': newPassword,
        'confirm_password': confirmPassword,
      },
    );

    if (_currentUser != null) {
      _currentUser = UserModel(
        id: _currentUser!.id,
        name: _currentUser!.name,
        email: _currentUser!.email,
        mustResetPassword: false,
        createdAt: _currentUser!.createdAt,
      );
      final prefs = await SharedPreferences.getInstance();
      await prefs.setString(_keyUser, jsonEncode(_currentUser!.toJson()));
      notifyListeners();
    }

    return data['message'] ?? 'Password reset successfully.';
  }

  Future<String> changePassword({
    required String currentPassword,
    required String newPassword,
    required String confirmPassword,
  }) async {
    final data = await _apiClient.post(
      ApiConstants.changePassword,
      data: {
        'current_password': currentPassword,
        'new_password': newPassword,
        'confirm_password': confirmPassword,
      },
    );
    return data['message'] ?? 'Password changed successfully.';
  }

  Future<UserModel> fetchMe() async {
    final data = await _apiClient.get(ApiConstants.me);
    final user = UserModel.fromJson(data);
    _currentUser = user;
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_keyUser, jsonEncode(user.toJson()));
    notifyListeners();
    return user;
  }

  Future<void> logout() async {
    try {
      if (_apiClient.authToken != null) {
        await _apiClient.post(ApiConstants.logout);
      }
    } catch (_) {}

    _apiClient.setAuthToken(null);
    _currentUser = null;
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_keyToken);
    await prefs.remove(_keyUser);
    notifyListeners();
  }

  Future<void> _saveAuth(String token, UserModel user) async {
    _apiClient.setAuthToken(token);
    _currentUser = user;
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_keyToken, token);
    await prefs.setString(_keyUser, jsonEncode(user.toJson()));
    notifyListeners();
  }
}
