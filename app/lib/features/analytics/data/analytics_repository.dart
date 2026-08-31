import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/providers/core_providers.dart';
import 'analytics_models.dart';

class AnalyticsRepository {
  AnalyticsRepository(this._dio);
  final Dio _dio;

  Future<OverallSummary> overallSummary({int? month, int? year}) async {
    final response = await _dio.get('/analytics/overall', queryParameters: {
      if (month != null) 'month': month,
      if (year != null) 'year': year,
    });
    return OverallSummary.fromJson(response.data as Map<String, dynamic>);
  }

  Future<ActivitySummary> activitySummary(String activityId, {int? month, int? year}) async {
    final response = await _dio.get('/analytics/activity', queryParameters: {
      'activity_id': activityId,
      if (month != null) 'month': month,
      if (year != null) 'year': year,
    });
    return ActivitySummary.fromJson(response.data as Map<String, dynamic>);
  }

  Future<List<StudentAttendanceSummary>> perfectAttendance(String activityId, {int? month, int? year}) async {
    final response = await _dio.get('/analytics/perfect-attendance', queryParameters: {
      'activity_id': activityId,
      if (month != null) 'month': month,
      if (year != null) 'year': year,
    });
    return (response.data as List<dynamic>)
        .map((e) => StudentAttendanceSummary.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<List<StudentAttendanceSummary>> monthlyReport(String activityId, {int? month, int? year}) async {
    final response = await _dio.get('/analytics/monthly-report', queryParameters: {
      'activity_id': activityId,
      if (month != null) 'month': month,
      if (year != null) 'year': year,
    });
    return (response.data as List<dynamic>)
        .map((e) => StudentAttendanceSummary.fromJson(e as Map<String, dynamic>))
        .toList();
  }
}

final analyticsRepositoryProvider = Provider<AnalyticsRepository>((ref) {
  return AnalyticsRepository(ref.watch(apiClientProvider).dio);
});

final overallSummaryProvider = FutureProvider.autoDispose<OverallSummary>((ref) {
  return ref.watch(analyticsRepositoryProvider).overallSummary();
});
