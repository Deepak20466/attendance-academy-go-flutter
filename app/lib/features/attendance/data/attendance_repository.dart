import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/paged_result.dart';
import '../../../core/providers/core_providers.dart';

class StudentAttendanceMark {
  StudentAttendanceMark.fromJson(Map<String, dynamic> json)
      : studentId = json['student_id'] as String,
        status = json['status'] as String;

  final String studentId;
  final String status;
}

class MarkEntry {
  MarkEntry({required this.studentId, required this.status, this.remarks = ''});
  final String studentId;
  final String status;
  final String remarks;

  Map<String, dynamic> toJson() => {'student_id': studentId, 'status': status, 'remarks': remarks};
}

class CoachAttendanceResult {
  CoachAttendanceResult.fromJson(Map<String, dynamic> json)
      : checkInVerified = json['check_in_verified'] as bool? ?? false;
  final bool checkInVerified;
}

class CoachAttendanceRecord {
  CoachAttendanceRecord.fromJson(Map<String, dynamic> json)
      : id = json['id'] as String,
        coachId = json['coach_id'] as String,
        classId = json['class_id'] as String?,
        attendanceDate = DateTime.parse(json['attendance_date'] as String),
        checkInTime = json['check_in_time'] == null ? null : DateTime.parse(json['check_in_time'] as String).toLocal(),
        checkInVerified = json['check_in_verified'] as bool,
        checkOutTime = json['check_out_time'] == null ? null : DateTime.parse(json['check_out_time'] as String).toLocal(),
        checkOutVerified = json['check_out_verified'] as bool,
        status = json['status'] as String;

  final String id;
  final String coachId;
  final String? classId;
  final DateTime attendanceDate;
  final DateTime? checkInTime;
  final bool checkInVerified;
  final DateTime? checkOutTime;
  final bool checkOutVerified;
  final String status;
}

class AttendanceRepository {
  AttendanceRepository(this._dio);
  final Dio _dio;

  Future<List<StudentAttendanceMark>> forClass(String classId) async {
    final response = await _dio.get('/attendance/students/class/$classId');
    return (response.data as List<dynamic>)
        .map((e) => StudentAttendanceMark.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<void> markBulk(String classId, List<MarkEntry> entries) async {
    await _dio.post('/attendance/students/mark', data: {
      'class_id': classId,
      'entries': entries.map((e) => e.toJson()).toList(),
    });
  }

  Future<CoachAttendanceResult> checkIn(String classId, double lat, double lng) async {
    final response = await _dio.post('/attendance/coaches/check-in', data: {
      'class_id': classId,
      'latitude': lat,
      'longitude': lng,
    });
    return CoachAttendanceResult.fromJson(response.data as Map<String, dynamic>);
  }

  Future<void> checkOut(String classId, double lat, double lng) async {
    await _dio.post('/attendance/coaches/check-out', data: {
      'class_id': classId,
      'latitude': lat,
      'longitude': lng,
    });
  }

  Future<PagedResult<CoachAttendanceRecord>> coachHistory({
    String? coachId,
    DateTime? dateFrom,
    DateTime? dateTo,
    int page = 1,
    int pageSize = 30,
  }) async {
    final response = await _dio.get('/attendance/coaches', queryParameters: {
      if (coachId != null) 'coach_id': coachId,
      if (dateFrom != null) 'date_from': _fmt(dateFrom),
      if (dateTo != null) 'date_to': _fmt(dateTo),
      'page': page,
      'page_size': pageSize,
    });
    return PagedResult.fromJson(response.data as Map<String, dynamic>, CoachAttendanceRecord.fromJson);
  }

  String _fmt(DateTime d) =>
      '${d.year.toString().padLeft(4, '0')}-${d.month.toString().padLeft(2, '0')}-${d.day.toString().padLeft(2, '0')}';
}

final attendanceRepositoryProvider = Provider<AttendanceRepository>((ref) {
  return AttendanceRepository(ref.watch(apiClientProvider).dio);
});
