import 'dart:async';

import 'package:dio/dio.dart';

import '../config/app_config.dart';
import '../storage/token_storage.dart';

/// Outcome of an attempted token refresh. [stale] is distinct from
/// [failed]: it means a *different, newer* session is now active (the user
/// logged out and back in while this refresh was in flight), so the right
/// response is to quietly drop this old request rather than log the new
/// session out.
enum _RefreshOutcome { refreshed, failed, stale }

/// Thin wrapper around Dio that attaches the access token to every request
/// and transparently refreshes it once on a 401 before retrying — the
/// mobile app never has to think about token lifecycle beyond calling
/// [logoutStream] to react to a session that couldn't be recovered.
class ApiClient {
  ApiClient(this._tokenStorage) {
    _dio = Dio(BaseOptions(
      baseUrl: AppConfig.apiBaseUrl,
      connectTimeout: const Duration(seconds: 15),
      receiveTimeout: const Duration(seconds: 15),
    ));

    _refreshDio = Dio(BaseOptions(baseUrl: AppConfig.apiBaseUrl));

    _dio.interceptors.add(InterceptorsWrapper(
      onRequest: (options, handler) async {
        final token = await _tokenStorage.readAccessToken();
        if (token != null) {
          options.headers['Authorization'] = 'Bearer $token';
        }
        handler.next(options);
      },
      onError: (error, handler) async {
        if (error.response?.statusCode == 401 && error.requestOptions.extra['retried'] != true) {
          final outcome = await _tryRefresh();
          switch (outcome) {
            case _RefreshOutcome.refreshed:
              final opts = error.requestOptions;
              opts.extra['retried'] = true;
              final token = await _tokenStorage.readAccessToken();
              opts.headers['Authorization'] = 'Bearer $token';
              try {
                final response = await _dio.fetch(opts);
                return handler.resolve(response);
              } catch (e) {
                return handler.next(error);
              }
            case _RefreshOutcome.failed:
              await _tokenStorage.clear();
              _loggedOutController.add(null);
            case _RefreshOutcome.stale:
              // A newer session is already active — this request belongs to
              // a session that's gone. Drop it without touching the current
              // session's tokens or forcing a logout.
              break;
          }
        }
        handler.next(error);
      },
    ));
  }

  late final Dio _dio;
  late final Dio _refreshDio;
  final TokenStorage _tokenStorage;

  final _loggedOutController = StreamController<void>.broadcast();
  Stream<void> get loggedOutStream => _loggedOutController.stream;

  Dio get dio => _dio;

  Future<_RefreshOutcome> _tryRefresh() async {
    final refreshToken = await _tokenStorage.readRefreshToken();
    if (refreshToken == null) return _RefreshOutcome.failed;
    try {
      final response = await _refreshDio.post('/auth/refresh', data: {'refresh_token': refreshToken});
      // A request made under a previous session can still be in flight —
      // retrying, hitting this 401 handler, and completing its refresh —
      // after the user has since logged out and logged back in as someone
      // else. Without this check, that stale refresh's response would land
      // here late and silently overwrite the new session's tokens with a
      // refreshed copy of the old one. Only commit if the refresh token we
      // started with is still the one on record.
      if (await _tokenStorage.readRefreshToken() != refreshToken) return _RefreshOutcome.stale;
      final data = response.data as Map<String, dynamic>;
      await _tokenStorage.saveTokens(
        accessToken: data['access_token'] as String,
        refreshToken: data['refresh_token'] as String,
      );
      return _RefreshOutcome.refreshed;
    } catch (_) {
      return _RefreshOutcome.failed;
    }
  }

  void dispose() {
    _loggedOutController.close();
  }
}
