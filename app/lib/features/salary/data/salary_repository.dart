import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/paged_result.dart';
import '../../../core/providers/core_providers.dart';
import 'salary_models.dart';

class SalaryRepository {
  SalaryRepository(this._dio);
  final Dio _dio;

  Future<PagedResult<SalaryAckModel>> mine({int page = 1}) async {
    final response = await _dio.get('/salary/mine', queryParameters: {'page': page});
    return PagedResult.fromJson(response.data as Map<String, dynamic>, SalaryAckModel.fromJson);
  }

  Future<PagedResult<SalaryAckModel>> forPeriod(int month, int year, {String? status, int page = 1}) async {
    final response = await _dio.get('/salary', queryParameters: {
      'month': month,
      'year': year,
      if (status != null) 'status': status,
      'page': page,
    });
    return PagedResult.fromJson(response.data as Map<String, dynamic>, SalaryAckModel.fromJson);
  }

  Future<void> acknowledge(String id) => _dio.post('/salary/$id/acknowledge');

  Future<int> generate(int month, int year) async {
    final response = await _dio.post('/salary/generate', data: {'month': month, 'year': year});
    return (response.data as Map<String, dynamic>)['created'] as int;
  }
}

final salaryRepositoryProvider = Provider<SalaryRepository>((ref) {
  return SalaryRepository(ref.watch(apiClientProvider).dio);
});
