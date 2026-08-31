import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/paged_result.dart';
import '../../../core/providers/core_providers.dart';
import 'fee_models.dart';

class FeeRepository {
  FeeRepository(this._dio);
  final Dio _dio;

  Future<PagedResult<FeeModel>> list({String? status, int page = 1}) async {
    final response = await _dio.get('/fees', queryParameters: {
      if (status != null) 'status': status,
      'page': page,
    });
    return PagedResult.fromJson(response.data as Map<String, dynamic>, FeeModel.fromJson);
  }

  Future<void> markPaid(String id, {String paymentMethod = ''}) async {
    await _dio.post('/fees/$id/paid', data: {'payment_method': paymentMethod});
  }

  Future<int> generate({
    required String activityId,
    required double amount,
    required DateTime dueDate,
    required int month,
    required int year,
  }) async {
    final response = await _dio.post('/fees/generate', data: {
      'activity_id': activityId,
      'amount': amount,
      'due_date':
          '${dueDate.year.toString().padLeft(4, '0')}-${dueDate.month.toString().padLeft(2, '0')}-${dueDate.day.toString().padLeft(2, '0')}',
      'period_month': month,
      'period_year': year,
    });
    return (response.data as Map<String, dynamic>)['created'] as int;
  }
}

final feeRepositoryProvider = Provider<FeeRepository>((ref) {
  return FeeRepository(ref.watch(apiClientProvider).dio);
});
