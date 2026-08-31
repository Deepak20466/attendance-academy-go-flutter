import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/paged_result.dart';
import '../../../core/providers/core_providers.dart';
import 'student_models.dart';

class AttendanceHistoryEntry {
  AttendanceHistoryEntry.fromJson(Map<String, dynamic> json)
    : classId = json['class_id'] as String,
      status = json['status'] as String,
      // .toLocal(): a real timestamp, not a pure calendar date — must
      // render in the viewer's timezone (see notification_repository.dart
      // for why this matters).
      markedAt = DateTime.parse(json['marked_at'] as String).toLocal(),
      remarks = json['remarks'] as String? ?? '';

  final String classId;
  final String status;
  final DateTime markedAt;
  final String remarks;
}

class MonthlyPercentage {
  MonthlyPercentage.fromJson(Map<String, dynamic> json)
    : totalClasses = json['total_classes'] as int,
      presentCount = json['present_count'] as int,
      absentCount = json['absent_count'] as int,
      percentPresent = (json['percent_present'] as num).toDouble();

  final int totalClasses;
  final int presentCount;
  final int absentCount;
  final double percentPresent;
}

class StudentRepository {
  StudentRepository(this._dio);
  final Dio _dio;

  Future<PagedResult<StudentModel>> list({
    String? activityId,
    String? batchId,
    int page = 1,
    int pageSize = 30,
  }) async {
    final response = await _dio.get(
      '/students',
      queryParameters: {
        if (activityId != null) 'activity_id': activityId,
        if (batchId != null) 'batch_id': batchId,
        'page': page,
        'page_size': pageSize,
      },
    );
    return PagedResult.fromJson(
      response.data as Map<String, dynamic>,
      StudentModel.fromJson,
    );
  }

  Future<StudentModel> get(String id) async {
    final response = await _dio.get('/students/$id');
    return StudentModel.fromJson(response.data as Map<String, dynamic>);
  }

  Future<void> create({
    required String batchId,
    required String name,
    String phone = '',
    String guardianName = '',
    String guardianPhone = '',
    String email = '',
  }) async {
    await _dio.post(
      '/students',
      data: {
        'batch_id': batchId,
        'name': name,
        'phone': phone,
        'guardian_name': guardianName,
        'guardian_phone': guardianPhone,
        'email': email,
      },
    );
  }

  // The backend has no separate deactivate/delete endpoint — it overwrites
  // every column on PUT, so callers (edit AND remove) must always pass the
  // full current record back, including dateOfBirth, or they'll silently
  // wipe fields they never meant to touch.
  Future<void> update({
    required String id,
    required String batchId,
    required String name,
    String phone = '',
    String guardianName = '',
    String guardianPhone = '',
    String email = '',
    DateTime? dateOfBirth,
    bool isActive = true,
  }) async {
    await _dio.put(
      '/students/$id',
      data: {
        'batch_id': batchId,
        'name': name,
        'phone': phone,
        'guardian_name': guardianName,
        'guardian_phone': guardianPhone,
        'email': email,
        if (dateOfBirth != null) 'date_of_birth': dateOfBirth.toIso8601String(),
        'is_active': isActive,
      },
    );
  }

  Future<PagedResult<AttendanceHistoryEntry>> history(
    String studentId, {
    int page = 1,
    int pageSize = 30,
  }) async {
    final response = await _dio.get(
      '/attendance/students/student/$studentId/history',
      queryParameters: {'page': page, 'page_size': pageSize},
    );
    return PagedResult.fromJson(
      response.data as Map<String, dynamic>,
      AttendanceHistoryEntry.fromJson,
    );
  }

  Future<MonthlyPercentage> monthlyPercentage(
    String studentId,
    String batchId,
    int year,
    int month,
  ) async {
    final response = await _dio.get(
      '/attendance/students/student/$studentId/monthly',
      queryParameters: {'batch_id': batchId, 'year': year, 'month': month},
    );
    return MonthlyPercentage.fromJson(response.data as Map<String, dynamic>);
  }
}

final studentRepositoryProvider = Provider<StudentRepository>((ref) {
  return StudentRepository(ref.watch(apiClientProvider).dio);
});

final studentsListProvider =
    FutureProvider.autoDispose<PagedResult<StudentModel>>((ref) {
      return ref.watch(studentRepositoryProvider).list();
    });

final studentDetailProvider = FutureProvider.autoDispose
    .family<StudentModel, String>((ref, id) {
      return ref.watch(studentRepositoryProvider).get(id);
    });
