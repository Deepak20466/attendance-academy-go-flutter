import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/paged_result.dart';
import '../../../core/providers/core_providers.dart';
import 'class_models.dart';

class ClassRepository {
  ClassRepository(this._dio);
  final Dio _dio;

  Future<PagedResult<ClassModel>> list({
    String? activityId,
    String? coachId,
    DateTime? dateFrom,
    DateTime? dateTo,
    int page = 1,
    int pageSize = 50,
  }) async {
    final response = await _dio.get('/classes', queryParameters: {
      if (activityId != null) 'activity_id': activityId,
      if (coachId != null) 'coach_id': coachId,
      if (dateFrom != null) 'date_from': _fmt(dateFrom),
      if (dateTo != null) 'date_to': _fmt(dateTo),
      'page': page,
      'page_size': pageSize,
    });
    return PagedResult.fromJson(response.data as Map<String, dynamic>, ClassModel.fromJson);
  }

  Future<ClassModel> get(String id) async {
    final response = await _dio.get('/classes/$id');
    return ClassModel.fromJson(response.data as Map<String, dynamic>);
  }

  // Deliberately class-scoped, not activity-scoped: a substitute coach
  // with no general activity membership is still authorized for the one
  // class they're covering, and needs its roster to mark attendance.
  Future<List<RosterStudent>> roster(String classId) async {
    final response = await _dio.get('/classes/$classId/roster');
    return (response.data as List<dynamic>).map((e) => RosterStudent.fromJson(e as Map<String, dynamic>)).toList();
  }

  Future<List<BatchModel>> listBatches({String? activityId}) async {
    final response = await _dio.get('/batches', queryParameters: {
      if (activityId != null) 'activity_id': activityId,
    });
    return (response.data as List<dynamic>).map((e) => BatchModel.fromJson(e as Map<String, dynamic>)).toList();
  }

  Future<void> createBatch({
    required String activityId,
    required String name,
    required String description,
    String? defaultCoachId,
    String? locationId,
    required List<int> daysOfWeek,
    required String startTime,
    required String endTime,
  }) async {
    await _dio.post('/batches', data: {
      'activity_id': activityId,
      'name': name,
      'description': description,
      'default_coach_id': defaultCoachId,
      'location_id': locationId,
      'days_of_week': daysOfWeek,
      'start_time': startTime,
      'end_time': endTime,
    });
  }

  Future<void> createClass({
    required String batchId,
    required String activityId,
    required String coachId,
    required DateTime classDate,
    required String startTime,
    required String endTime,
  }) async {
    await _dio.post('/classes', data: {
      'batch_id': batchId,
      'activity_id': activityId,
      'coach_id': coachId,
      'class_date': _fmt(classDate),
      'start_time': startTime,
      'end_time': endTime,
    });
  }

  String _fmt(DateTime d) =>
      '${d.year.toString().padLeft(4, '0')}-${d.month.toString().padLeft(2, '0')}-${d.day.toString().padLeft(2, '0')}';
}

final classRepositoryProvider = Provider<ClassRepository>((ref) {
  return ClassRepository(ref.watch(apiClientProvider).dio);
});

final todayClassesProvider = FutureProvider.autoDispose<List<ClassModel>>((ref) async {
  final repo = ref.watch(classRepositoryProvider);
  final now = DateTime.now();
  final result = await repo.list(dateFrom: now, dateTo: now, pageSize: 100);
  return result.data;
});

// Exposed here (rather than private to ClassesScreen) so other screens —
// notably the attendance-marking flow — can invalidate it after a mutation.
// This matters because ClassesScreen lives inside HomeShell's IndexedStack
// and never unmounts on tab switches, so autoDispose alone never triggers
// a refetch after returning from marking attendance.
final recentClassesProvider = FutureProvider.autoDispose<PagedResult<ClassModel>>((ref) {
  final repo = ref.watch(classRepositoryProvider);
  final now = DateTime.now();
  final from = now.subtract(const Duration(days: 14));
  return repo.list(dateFrom: from, dateTo: now, pageSize: 50);
});
