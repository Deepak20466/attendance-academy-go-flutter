class AdminUserModel {
  AdminUserModel.fromJson(Map<String, dynamic> json)
      : id = json['id'] as String,
        role = json['role'] as String,
        name = json['name'] as String,
        email = json['email'] as String,
        phone = json['phone'] as String? ?? '',
        isActive = json['is_active'] as bool,
        createdAt = DateTime.parse(json['created_at'] as String).toLocal();

  final String id;
  final String role;
  final String name;
  final String email;
  final String phone;
  final bool isActive;
  final DateTime createdAt;
}
