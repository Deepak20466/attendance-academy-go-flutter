import 'package:dio/dio.dart';

import '../../../core/storage/token_storage.dart';
import 'auth_models.dart';

class AuthRepository {
  AuthRepository(this._dio, this._tokenStorage);

  final Dio _dio;
  final TokenStorage _tokenStorage;

  Future<AuthSession> login(String email, String password) async {
    final response = await _dio.post('/auth/login', data: {
      'email': email,
      'password': password,
    });
    final data = response.data as Map<String, dynamic>;
    await _tokenStorage.saveTokens(
      accessToken: data['access_token'] as String,
      refreshToken: data['refresh_token'] as String,
    );
    return AuthSession.fromLoginJson(data);
  }

  Future<AuthSession?> restoreSession() async {
    final token = await _tokenStorage.readAccessToken();
    if (token == null) return null;
    try {
      final response = await _dio.get('/auth/me');
      final data = response.data as Map<String, dynamic>;
      // /auth/me does not return name/email; fall back to placeholders that
      // get refreshed the next time the user logs in explicitly.
      return AuthSession(
        userId: data['user_id'] as String,
        name: '',
        email: '',
        role: data['role'] as String,
      );
    } catch (_) {
      return null;
    }
  }

  Future<void> logout() async {
    final refreshToken = await _tokenStorage.readRefreshToken();
    if (refreshToken != null) {
      try {
        await _dio.post('/auth/logout', data: {'refresh_token': refreshToken});
      } catch (_) {
        // best-effort — clear local session regardless
      }
    }
    await _tokenStorage.clear();
  }

  Future<void> updateFcmToken(String fcmToken) async {
    await _dio.put('/auth/fcm-token', data: {'fcm_token': fcmToken});
  }
}
