class StudentModel {
  StudentModel.fromJson(Map<String, dynamic> json)
      : id = json['id'] as String,
        batchId = json['batch_id'] as String,
        activityId = json['activity_id'] as String,
        name = json['name'] as String,
        phone = json['phone'] as String? ?? '',
        guardianName = json['guardian_name'] as String? ?? '',
        guardianPhone = json['guardian_phone'] as String? ?? '',
        email = json['email'] as String? ?? '',
        isActive = json['is_active'] as bool;

  final String id;
  final String batchId;
  final String activityId;
  final String name;
  final String phone;
  final String guardianName;
  final String guardianPhone;
  final String email;
  final bool isActive;
}
