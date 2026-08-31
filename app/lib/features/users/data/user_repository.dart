import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/paged_result.dart';
import '../../../core/providers/core_providers.dart';
import 'user_models.dart';

class UserRepository {
  UserRepository(this._dio);
  final Dio _dio;

  Future<PagedResult<AdminUserModel>> list({String? role, int page = 1, int pageSize = 50}) async {
    final response = await _dio.get('/users', queryParameters: {
      if (role != null) 'role': role,
      'page': page,
      'page_size': pageSize,
    });
    return PagedResult.fromJson(response.data as Map<String, dynamic>, AdminUserModel.fromJson);
  }

  Future<void> createAdmin({
    required String name,
    required String email,
    required String phone,
    required String password,
  }) async {
    await _dio.post('/users', data: {
      'name': name,
      'email': email,
      'phone': phone,
      'password': password,
    });
  }

  Future<void> setActive(String id, bool active) async {
    await _dio.post('/users/$id/${active ? 'activate' : 'deactivate'}');
  }
}

final userRepositoryProvider = Provider<UserRepository>((ref) {
  return UserRepository(ref.watch(apiClientProvider).dio);
});
