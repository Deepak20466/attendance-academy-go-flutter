class ClassModel {
  ClassModel.fromJson(Map<String, dynamic> json)
      : id = json['id'] as String,
        batchId = json['batch_id'] as String,
        activityId = json['activity_id'] as String,
        coachId = json['coach_id'] as String,
        classDate = DateTime.parse(json['class_date'] as String),
        startTime = json['start_time'] as String,
        endTime = json['end_time'] as String,
        status = json['status'] as String,
        attendanceMarked = json['attendance_marked'] as bool;

  final String id;
  final String batchId;
  final String activityId;
  final String coachId;
  final DateTime classDate;
  final String startTime;
  final String endTime;
  final String status;
  final bool attendanceMarked;

  bool get isToday {
    final now = DateTime.now();
    return classDate.year == now.year && classDate.month == now.month && classDate.day == now.day;
  }
}

class RosterStudent {
  RosterStudent.fromJson(Map<String, dynamic> json)
      : id = json['id'] as String,
        name = json['name'] as String;

  final String id;
  final String name;
}

class BatchModel {
  BatchModel.fromJson(Map<String, dynamic> json)
      : id = json['id'] as String,
        activityId = json['activity_id'] as String,
        name = json['name'] as String,
        defaultCoachId = json['default_coach_id'] as String?,
        startTime = json['start_time'] as String,
        endTime = json['end_time'] as String,
        daysOfWeek = (json['days_of_week'] as List<dynamic>? ?? []).cast<int>(),
        isActive = json['is_active'] as bool;

  final String id;
  final String activityId;
  final String name;
  final String? defaultCoachId;
  final String startTime;
  final String endTime;
  final List<int> daysOfWeek;
  final bool isActive;
}
