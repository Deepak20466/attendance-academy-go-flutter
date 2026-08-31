class FeeModel {
  FeeModel.fromJson(Map<String, dynamic> json)
      : id = json['id'] as String,
        studentId = json['student_id'] as String,
        activityId = json['activity_id'] as String,
        amount = (json['amount'] as num).toDouble(),
        dueDate = DateTime.parse(json['due_date'] as String),
        status = json['status'] as String,
        periodMonth = json['period_month'] as int,
        periodYear = json['period_year'] as int;

  final String id;
  final String studentId;
  final String activityId;
  final double amount;
  final DateTime dueDate;
  final String status;
  final int periodMonth;
  final int periodYear;
}
