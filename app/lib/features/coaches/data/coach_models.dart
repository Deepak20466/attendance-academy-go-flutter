class CoachModel {
  CoachModel.fromJson(Map<String, dynamic> json)
      : id = json['id'] as String,
        userId = json['user_id'] as String,
        name = json['name'] as String,
        email = json['email'] as String,
        phone = json['phone'] as String? ?? '',
        employeeCode = json['employee_code'] as String,
        monthlySalary = (json['monthly_salary'] as num).toDouble(),
        isActive = json['is_active'] as bool,
        activityIds = (json['activity_ids'] as List<dynamic>? ?? []).cast<String>();

  final String id;
  final String userId;
  final String name;
  final String email;
  final String phone;
  final String employeeCode;
  final double monthlySalary;
  final bool isActive;
  final List<String> activityIds;
}
