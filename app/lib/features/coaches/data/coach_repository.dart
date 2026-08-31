import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/paged_result.dart';
import '../../../core/providers/core_providers.dart';
import 'coach_models.dart';

class CoachRepository {
  CoachRepository(this._dio);
  final Dio _dio;

  Future<CoachModel> me() async {
    final response = await _dio.get('/coaches/me');
    return CoachModel.fromJson(response.data as Map<String, dynamic>);
  }

  Future<CoachModel> get(String id) async {
    final response = await _dio.get('/coaches/$id');
    return CoachModel.fromJson(response.data as Map<String, dynamic>);
  }

  Future<PagedResult<CoachModel>> list({String? activityId, int page = 1, int pageSize = 50}) async {
    final response = await _dio.get('/coaches', queryParameters: {
      if (activityId != null) 'activity_id': activityId,
      'page': page,
      'page_size': pageSize,
    });
    return PagedResult.fromJson(response.data as Map<String, dynamic>, CoachModel.fromJson);
  }

  Future<void> create({
    required String name,
    required String email,
    required String phone,
    required String password,
    required String employeeCode,
    required double monthlySalary,
    required List<String> activityIds,
  }) async {
    await _dio.post('/coaches', data: {
      'name': name,
      'email': email,
      'phone': phone,
      'password': password,
      'employee_code': employeeCode,
      'monthly_salary': monthlySalary,
      'activity_ids': activityIds,
    });
  }
}

final coachRepositoryProvider = Provider<CoachRepository>((ref) {
  return CoachRepository(ref.watch(apiClientProvider).dio);
});

final myCoachProfileProvider = FutureProvider.autoDispose<CoachModel>((ref) {
  return ref.watch(coachRepositoryProvider).me();
});
