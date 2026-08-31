import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/paged_result.dart';
import '../../../core/providers/core_providers.dart';
import 'leave_models.dart';

class LeaveRepository {
  LeaveRepository(this._dio);
  final Dio _dio;

  Future<PagedResult<LeaveModel>> mine({int page = 1}) async {
    final response = await _dio.get('/leaves/mine', queryParameters: {'page': page});
    return PagedResult.fromJson(response.data as Map<String, dynamic>, LeaveModel.fromJson);
  }

  Future<PagedResult<LeaveModel>> all({String? status, int page = 1}) async {
    final response = await _dio.get('/leaves', queryParameters: {
      if (status != null) 'status': status,
      'page': page,
    });
    return PagedResult.fromJson(response.data as Map<String, dynamic>, LeaveModel.fromJson);
  }

  Future<void> apply(DateTime start, DateTime end, String reason) async {
    await _dio.post('/leaves', data: {
      'start_date': _fmt(start),
      'end_date': _fmt(end),
      'reason': reason,
    });
  }

  Future<void> approve(String id) => _dio.post('/leaves/$id/approve');
  Future<void> reject(String id) => _dio.post('/leaves/$id/reject');
  Future<void> cancel(String id) => _dio.post('/leaves/$id/cancel');

  String _fmt(DateTime d) =>
      '${d.year.toString().padLeft(4, '0')}-${d.month.toString().padLeft(2, '0')}-${d.day.toString().padLeft(2, '0')}';
}

final leaveRepositoryProvider = Provider<LeaveRepository>((ref) {
  return LeaveRepository(ref.watch(apiClientProvider).dio);
});

// Exposed here (rather than private to LeavesScreen) so LeaveDetailScreen
// can invalidate them after approve/reject/cancel — LeavesScreen lives
// inside HomeShell's IndexedStack and never unmounts, so autoDispose alone
// never triggers a refetch after returning from the detail screen.
final myLeavesProvider = FutureProvider.autoDispose((ref) {
  return ref.watch(leaveRepositoryProvider).mine();
});

final allLeavesProvider = FutureProvider.autoDispose((ref) {
  return ref.watch(leaveRepositoryProvider).all();
});
