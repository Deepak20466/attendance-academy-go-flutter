class OverallSummary {
  OverallSummary.fromJson(Map<String, dynamic> json)
      : totalStudents = json['total_students'] as int,
        totalCoaches = json['total_coaches'] as int,
        totalActivities = json['total_activities'] as int,
        totalClassesThisMonth = json['total_classes_this_month'] as int,
        classesWithAttendance = json['classes_with_attendance'] as int,
        attendancePercent = (json['attendance_percent'] as num).toDouble(),
        feesCollectedThisMonth = (json['fees_collected_this_month'] as num).toDouble(),
        feesPendingThisMonth = (json['fees_pending_this_month'] as num).toDouble(),
        pendingLeaves = json['pending_leaves'] as int,
        coachCheckInsToday = json['coach_check_ins_today'] as int,
        todayClasses = json['today_classes'] as int,
        todayMissingAttendance = json['today_missing_attendance'] as int;

  final int totalStudents;
  final int totalCoaches;
  final int totalActivities;
  final int totalClassesThisMonth;
  final int classesWithAttendance;
  final double attendancePercent;
  final double feesCollectedThisMonth;
  final double feesPendingThisMonth;
  final int pendingLeaves;
  final int coachCheckInsToday;
  final int todayClasses;
  final int todayMissingAttendance;
}

class ActivitySummary {
  ActivitySummary.fromJson(Map<String, dynamic> json)
      : activityId = json['activity_id'] as String,
        activityName = json['activity_name'] as String,
        studentCount = json['student_count'] as int,
        coachCount = json['coach_count'] as int,
        classCount = json['class_count'] as int,
        presentCount = json['present_count'] as int,
        absentCount = json['absent_count'] as int,
        attendancePercent = (json['attendance_percent'] as num).toDouble(),
        perfectAttendance = json['perfect_attendance_students'] as int,
        feesCollected = (json['fees_collected'] as num).toDouble(),
        feesPending = (json['fees_pending'] as num).toDouble(),
        coachAttendanceDays = json['coach_attendance_days'] as int;

  final String activityId;
  final String activityName;
  final int studentCount;
  final int coachCount;
  final int classCount;
  final int presentCount;
  final int absentCount;
  final double attendancePercent;
  final int perfectAttendance;
  final double feesCollected;
  final double feesPending;
  final int coachAttendanceDays;
}

class StudentAttendanceSummary {
  StudentAttendanceSummary.fromJson(Map<String, dynamic> json)
      : studentId = json['student_id'] as String,
        studentName = json['student_name'] as String,
        totalClasses = json['total_classes'] as int,
        presentCount = json['present_count'] as int,
        percentPresent = (json['percent_present'] as num).toDouble();

  final String studentId;
  final String studentName;
  final int totalClasses;
  final int presentCount;
  final double percentPresent;
}
