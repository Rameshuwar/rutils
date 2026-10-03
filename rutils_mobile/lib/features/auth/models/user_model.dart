class UserModel {
  final String id;
  final String name;
  final String email;
  final bool mustResetPassword;
  final String? createdAt;

  UserModel({
    required this.id,
    required this.name,
    required this.email,
    required this.mustResetPassword,
    this.createdAt,
  });

  factory UserModel.fromJson(Map<String, dynamic> json) {
    return UserModel(
      id: json['id'] ?? '',
      name: json['name'] ?? '',
      email: json['email'] ?? '',
      mustResetPassword: json['must_reset_password'] ?? false,
      createdAt: json['created_at'],
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'name': name,
      'email': email,
      'must_reset_password': mustResetPassword,
      'created_at': createdAt,
    };
  }
}

class LoginResponse {
  final String token;
  final UserModel user;
  final bool mustResetPassword;
  final String message;

  LoginResponse({
    required this.token,
    required this.user,
    required this.mustResetPassword,
    required this.message,
  });

  factory LoginResponse.fromJson(Map<String, dynamic> json) {
    return LoginResponse(
      token: json['token'] ?? '',
      user: UserModel.fromJson(json['user'] ?? {}),
      mustResetPassword: json['must_reset_password'] ?? false,
      message: json['message'] ?? '',
    );
  }
}
